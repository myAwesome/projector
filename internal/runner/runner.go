package runner

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"proj/internal/store"
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
	dir := filepath.Join(home, ".config", "proj")
	return dir, os.MkdirAll(dir, 0755)
}

func stateFile() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "state.json"), nil
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

	if err := cmd.Start(); err != nil {
		return RunState{}, err
	}

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
