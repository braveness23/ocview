package data

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var knownHookDescriptions = map[string]string{
	"boot-md":               "Load workspace markdown files on session boot",
	"bootstrap-extra-files": "Inject extra context files into session startup",
	"command-logger":        "Log all executed commands to audit trail",
	"session-memory":        "Save session interactions to long-term memory",
}

func LoadHooks() []OcHook {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".openclaw", "openclaw.json")

	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var cfg openclawJSON
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil
	}
	if cfg.Hooks == nil || cfg.Hooks.Internal == nil {
		return nil
	}

	var items []OcHook
	for name, entry := range cfg.Hooks.Internal.Entries {
		enabled := true
		if v, ok := entry["enabled"]; ok {
			if b, ok := v.(bool); ok {
				enabled = b
			}
		}
		items = append(items, OcHook{
			ID:          "hook#" + name,
			Name_:       name,
			Description: knownHookDescriptions[name],
			Enabled:     enabled,
			RawConfig:   entry,
		})
	}
	return items
}
