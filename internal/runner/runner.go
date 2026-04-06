package runner

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"project/internal/store"
)

type RunState struct {
	PID       int       `json:"pid"`
	PGID      int       `json:"pgid"`
	StartedAt time.Time `json:"started_at"`
}

var ErrAlreadyRunning = errors.New("project is already running")
var ErrNotRunning = errors.New("project is not running")

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "project")
	return dir, os.MkdirAll(dir, 0755)
}

func stateFile() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "state.json"), nil
}

func logsDir() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	logs := filepath.Join(dir, "logs")
	return logs, os.MkdirAll(logs, 0755)
}

func logFile(name string) (string, error) {
	dir, err := logsDir()
	if err != nil {
		return "", err
	}
	safeName := strings.ReplaceAll(strings.TrimSpace(name), string(filepath.Separator), "_")
	if safeName == "" {
		safeName = "project"
	}
	return filepath.Join(dir, safeName+".log"), nil
}

func loadState() (map[string]RunState, error) {
	path, err := stateFile()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]RunState{}, nil
	}
	if err != nil {
		return nil, err
	}
	var state map[string]RunState
	return state, json.Unmarshal(data, &state)
}

func saveState(state map[string]RunState) error {
	path, err := stateFile()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func IsRunning(name string) (RunState, bool) {
	state, err := loadState()
	if err != nil {
		return RunState{}, false
	}
	rs, ok := state[name]
	if !ok {
		return RunState{}, false
	}
	proc, err := os.FindProcess(rs.PID)
	if err != nil {
		return RunState{}, false
	}
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		// process is dead — clean up stale state
		delete(state, name)
		_ = saveState(state)
		return RunState{}, false
	}
	return rs, true
}

func Start(p store.Project) (RunState, error) {
	if _, running := IsRunning(p.Name); running {
		return RunState{}, ErrAlreadyRunning
	}

	cmd := exec.Command("sh", "-c", p.Script)
	cmd.Dir = p.Dir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	logPath, err := logFile(p.Name)
	if err != nil {
		return RunState{}, err
	}
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return RunState{}, err
	}
	_, _ = fmt.Fprintf(f, "\n[%s] starting %q in %s\n", time.Now().Format(time.RFC3339), p.Name, p.Dir)
	cmd.Stdout = f
	cmd.Stderr = f

	if err := cmd.Start(); err != nil {
		_ = f.Close()
		return RunState{}, err
	}
	_ = f.Close()

	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err != nil {
		pgid = cmd.Process.Pid
	}

	rs := RunState{
		PID:       cmd.Process.Pid,
		PGID:      pgid,
		StartedAt: time.Now(),
	}

	state, err := loadState()
	if err != nil {
		return RunState{}, err
	}
	state[p.Name] = rs
	return rs, saveState(state)
}

func Stop(name string) error {
	rs, running := IsRunning(name)
	if !running {
		return ErrNotRunning
	}

	if err := syscall.Kill(-rs.PGID, syscall.SIGTERM); err != nil {
		return err
	}

	state, err := loadState()
	if err != nil {
		return err
	}
	delete(state, name)
	return saveState(state)
}

func State(name string) (RunState, bool) {
	return IsRunning(name)
}

func AllStates() (map[string]RunState, error) {
	return loadState()
}

func TailLogs(name string, lines int) (string, error) {
	if lines <= 0 {
		lines = 80
	}
	path, err := logFile(name)
	if err != nil {
		return "", err
	}
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return "(no logs yet)", nil
	}
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// Allow long application log lines.
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)
	buf := make([]string, 0, lines)
	for scanner.Scan() {
		buf = append(buf, scanner.Text())
		if len(buf) > lines {
			buf = buf[1:]
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if len(buf) == 0 {
		return "(no logs yet)", nil
	}
	return strings.Join(buf, "\n"), nil
}
