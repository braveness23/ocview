package data

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func LoadLogs() []OcLogFile {
	home, _ := os.UserHomeDir()
	logsDir := filepath.Join(home, ".openclaw", "logs")

	var items []OcLogFile

	// journalctl — most useful first
	if out, err := exec.Command(
		"journalctl", "--user", "-u", "openclaw-gateway.service",
		"--no-pager", "-n", "300", "--output=short-iso",
	).Output(); err == nil {
		content := string(out)
		items = append(items, OcLogFile{
			ID:        "log#journal",
			Name_:     "journal (gateway)",
			FilePath:  "",
			Content:   content,
			LineCount: strings.Count(content, "\n") + 1,
		})
	}

	// sync-token.log
	if item := readLogTextFile(
		filepath.Join(logsDir, "sync-token.log"),
		"log#sync-token", "sync-token.log",
	); item != nil {
		items = append(items, *item)
	}

	// config-health.json (pretty-printed)
	healthPath := filepath.Join(logsDir, "config-health.json")
	if b, err := os.ReadFile(healthPath); err == nil {
		var raw any
		if json.Unmarshal(b, &raw) == nil {
			pretty, _ := json.MarshalIndent(raw, "", "  ")
			content := string(pretty)
			items = append(items, OcLogFile{
				ID:        "log#config-health",
				Name_:     "config-health.json",
				FilePath:  healthPath,
				Content:   content,
				LineCount: strings.Count(content, "\n") + 1,
			})
		}
	}

	return items
}

func readLogTextFile(path, id, name string) *OcLogFile {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	content := string(b)
	return &OcLogFile{
		ID:        id,
		Name_:     name,
		FilePath:  path,
		Content:   content,
		LineCount: strings.Count(content, "\n") + 1,
	}
}
