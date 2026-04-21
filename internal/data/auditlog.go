package data

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

func shortCommand(argv []string) string {
	for i, a := range argv {
		if strings.Contains(a, "openclaw") && !strings.HasPrefix(a, "--") {
			parts := argv[i:]
			var filtered []string
			for _, p := range parts {
				if !strings.HasPrefix(p, "--disable-warning") {
					filtered = append(filtered, p)
				}
			}
			return strings.Join(filtered, " ")
		}
	}
	if len(argv) > 1 {
		return strings.Join(argv[1:], " ")
	}
	return strings.Join(argv, " ")
}

func LoadAuditLog() []OcAuditEntry {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".openclaw", "logs", "config-audit.jsonl")

	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var entries []OcAuditEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var obj struct {
			TS                string   `json:"ts"`
			Event             string   `json:"event"`
			Source            string   `json:"source"`
			ConfigPath        string   `json:"configPath"`
			Argv              []string `json:"argv"`
			PID               int      `json:"pid"`
			Result            string   `json:"result"`
			Suspicious        []string `json:"suspicious"`
			PreviousBytes     *int64   `json:"previousBytes"`
			NextBytes         int64    `json:"nextBytes"`
			PreviousHash      string   `json:"previousHash"`
			NextHash          string   `json:"nextHash"`
			GatewayModeBefore *string  `json:"gatewayModeBefore"`
			GatewayModeAfter  *string  `json:"gatewayModeAfter"`
		}
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			continue
		}
		cmd := shortCommand(obj.Argv)
		entries = append(entries, OcAuditEntry{
			ID:                obj.TS,
			Name_:             cmd,
			TS:                obj.TS,
			Event:             obj.Event,
			Source:            obj.Source,
			ConfigPath:        obj.ConfigPath,
			Command:           cmd,
			Argv:              obj.Argv,
			PID:               obj.PID,
			Result:            obj.Result,
			Suspicious:        obj.Suspicious,
			PreviousBytes:     obj.PreviousBytes,
			NextBytes:         obj.NextBytes,
			PreviousHash:      obj.PreviousHash,
			NextHash:          obj.NextHash,
			GatewayModeBefore: obj.GatewayModeBefore,
			GatewayModeAfter:  obj.GatewayModeAfter,
		})
	}

	// Reverse for most-recent-first
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	return entries
}
