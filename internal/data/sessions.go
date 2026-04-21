package data

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var knownChannels = map[string]bool{
	"whatsapp": true, "telegram": true, "slack": true,
	"discord": true, "web": true, "voice": true, "sms": true,
}

func extractChannel(filePath string) string {
	stem := strings.TrimSuffix(filepath.Base(filePath), ".jsonl")
	if idx := strings.IndexByte(stem, '-'); idx > 0 {
		prefix := strings.ToLower(stem[:idx])
		if knownChannels[prefix] {
			return prefix
		}
	}

	f, err := os.Open(filePath)
	if err != nil {
		return "main"
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for i := 0; i < 8 && scanner.Scan(); i++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			continue
		}
		if ch, ok := obj["channel"].(string); ok && ch != "" {
			return ch
		}
		if meta, ok := obj["metadata"].(map[string]any); ok {
			if ch, ok := meta["channel"].(string); ok && ch != "" {
				return ch
			}
		}
		if sess, ok := obj["session"].(map[string]any); ok {
			if ch, ok := sess["channel"].(string); ok && ch != "" {
				return ch
			}
		}
	}
	return "main"
}

func LoadSessions() []OcSession {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".openclaw", "agents", "main", "sessions")

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var items []OcSession
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		filePath := filepath.Join(dir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}
		sessionID := strings.TrimSuffix(entry.Name(), ".jsonl")
		sizeKb := int(info.Size() / 1024)
		updatedAt := info.ModTime().UnixMilli()
		d := time.UnixMilli(updatedAt)
		label := fmt.Sprintf("%s %s  %s",
			d.Format("Jan"), fmt.Sprintf("%d", d.Day()),
			sessionID[:min(8, len(sessionID))],
		)

		items = append(items, OcSession{
			ID:          "session#" + sessionID,
			Name_:       label,
			Channel:     extractChannel(filePath),
			UpdatedAt:   updatedAt,
			SessionFile: filePath,
			SizeKb:      sizeKb,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].UpdatedAt > items[j].UpdatedAt
	})
	return items
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
