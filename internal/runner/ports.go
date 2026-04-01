package runner

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Ports returns TCP ports in LISTEN state for all processes in the given process group.
func Ports(pgid, pid int) []int {
	pids := pidsInGroup(pgid, pid)
	if len(pids) == 0 {
		return nil
	}

	// Build comma-separated PID list for lsof
	parts := make([]string, len(pids))
	for i, p := range pids {
		parts[i] = strconv.Itoa(p)
	}
	pidList := strings.Join(parts, ",")

	// -a: AND conditions; -p: filter by PIDs; -i TCP: internet TCP only; -P: numeric ports; -n: no hostname lookup
	out, err := exec.Command("lsof", "-a", "-p", pidList, "-i", "TCP", "-P", "-n").Output()
	if err != nil || len(out) == 0 {
		return nil
	}
	return parseLsofPorts(out)
}

// pidsInGroup returns all PIDs belonging to the process group (pgid).
// Falls back to just the root pid if pgrep is unavailable.
func pidsInGroup(pgid, pid int) []int {
	out, err := exec.Command("pgrep", "-g", strconv.Itoa(pgid)).Output()
	if err != nil || len(out) == 0 {
		return []int{pid}
	}
	var pids []int
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if p, err := strconv.Atoi(line); err == nil {
			pids = append(pids, p)
		}
	}
	if len(pids) == 0 {
		return []int{pid}
	}
	return pids
}

func parseLsofPorts(out []byte) []int {
	seen := map[int]bool{}
	var ports []int

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "LISTEN") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		// address field is second-to-last before "(LISTEN)"
		addr := fields[len(fields)-2]
		if i := strings.LastIndex(addr, ":"); i >= 0 {
			p, err := strconv.Atoi(addr[i+1:])
			if err == nil && !seen[p] {
				seen[p] = true
				ports = append(ports, p)
			}
		}
	}
	return ports
}

func FormatPorts(ports []int) string {
	if len(ports) == 0 {
		return "-"
	}
	parts := make([]string, len(ports))
	for i, p := range ports {
		parts[i] = fmt.Sprintf("%d", p)
	}
	return strings.Join(parts, ", ")
}
