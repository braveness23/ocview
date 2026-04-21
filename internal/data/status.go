package data

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func LoadStatus() ServiceStatus {
	var s ServiceStatus
	s.Active = "unknown"
	s.SocketHealth = "unknown"

	// Active state
	out, err := runCmd(2*time.Second, "systemctl", "--user", "is-active", "openclaw-gateway.service")
	if err != nil {
		s.Active = "stopped"
	} else {
		switch strings.TrimSpace(out) {
		case "active":
			s.Active = "running"
		case "failed":
			s.Active = "failed"
		default:
			s.Active = "stopped"
		}
	}

	// Since (only when running)
	if s.Active == "running" {
		out, err := runCmd(2*time.Second, "systemctl", "--user", "show",
			"openclaw-gateway.service", "--property=ActiveEnterTimestamp")
		if err == nil {
			if idx := strings.Index(out, "="); idx >= 0 {
				ts := strings.TrimSpace(out[idx+1:])
				if ts != "" && ts != "n/a" {
					if d, err := time.Parse("Mon 2006-01-02 15:04:05 MST", ts); err == nil {
						s.Since = d.Format("Jan 2, 3:04 PM")
					}
				}
			}
		}
	}

	// Socket health
	home, _ := os.UserHomeDir()
	healthPath := filepath.Join(home, ".openclaw", "logs", "config-health.json")
	if b, err := os.ReadFile(healthPath); err == nil {
		var h struct {
			SocketStatus string `json:"socketStatus"`
		}
		if json.Unmarshal(b, &h) == nil && h.SocketStatus != "" {
			if h.SocketStatus == "connected" {
				s.SocketHealth = "ok"
			} else {
				s.SocketHealth = "stale"
			}
		}
	}
	if s.SocketHealth == "unknown" && s.Active == "running" {
		logs, err := runCmd(2*time.Second, "journalctl", "--user",
			"-u", "openclaw-gateway.service", "-n", "20", "--no-pager", "--output=cat")
		if err == nil {
			if strings.Contains(logs, "stale-socket") {
				s.SocketHealth = "stale"
			} else {
				s.SocketHealth = "ok"
			}
		}
	}

	// Version
	npmOut, err := runCmd(2*time.Second, "npm", "root", "-g")
	if err == nil {
		pkgPath := filepath.Join(strings.TrimSpace(npmOut), "openclaw", "package.json")
		if b, err := os.ReadFile(pkgPath); err == nil {
			var pkg struct {
				Version string `json:"version"`
			}
			if json.Unmarshal(b, &pkg) == nil {
				s.Version = pkg.Version
			}
		}
	}

	return s
}

func runCmd(timeout time.Duration, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	ch := make(chan struct{ out string; err error }, 1)
	go func() {
		out, err := cmd.Output()
		ch <- struct{ out string; err error }{string(out), err}
	}()
	select {
	case r := <-ch:
		return r.out, r.err
	case <-time.After(timeout):
		cmd.Process.Kill()
		return "", os.ErrDeadlineExceeded
	}
}
