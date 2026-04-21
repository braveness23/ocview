package actions

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/braveness23/ocview/internal/data"
)

var openclawJSON = func() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".openclaw", "openclaw.json")
}()

var cronFile = func() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".openclaw", "cron", "jobs.json")
}()

var installedSkillsDir = func() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".openclaw", "skills")
}()

func GetEditableFilePath(item data.AnyItem) string {
	switch v := item.(type) {
	case data.OcSkill:
		return v.FilePath
	case data.OcWorkspaceFile:
		return v.FilePath
	case data.OcMemoryChunk:
		return v.Path
	}
	return ""
}

func EditorCmd(filePath string) *exec.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "nano"
	}
	return exec.Command(editor, filePath)
}

func CreateSkill(dirName string) (string, error) {
	skillDir := filepath.Join(installedSkillsDir, dirName)
	skillFile := filepath.Join(skillDir, "SKILL.md")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return "", err
	}
	content := "---\nname: " + dirName + "\ndescription: \n---\n\n# " + dirName + "\n"
	if err := os.WriteFile(skillFile, []byte(content), 0644); err != nil {
		return "", err
	}
	return skillFile, nil
}

func DeleteSkill(item data.OcSkill) error {
	dir := filepath.Dir(item.FilePath)
	return os.RemoveAll(dir)
}

func DeleteCronJob(item data.OcCronJob) (bool, error) {
	b, err := os.ReadFile(cronFile)
	if err != nil {
		return false, err
	}

	var raw any
	if err := json.Unmarshal(b, &raw); err != nil {
		return false, err
	}

	isArray := false
	var jobs []map[string]any

	switch v := raw.(type) {
	case []any:
		isArray = true
		for _, j := range v {
			if m, ok := j.(map[string]any); ok {
				jobs = append(jobs, m)
			}
		}
	case map[string]any:
		if js, ok := v["jobs"].([]any); ok {
			for _, j := range js {
				if m, ok := j.(map[string]any); ok {
					jobs = append(jobs, m)
				}
			}
		}
	}

	var filtered []map[string]any
	for _, j := range jobs {
		if j["id"] != item.ID && j["name"] != item.ItemName() {
			filtered = append(filtered, j)
		}
	}
	if len(filtered) == len(jobs) {
		return false, nil
	}

	var out any
	if isArray {
		out = filtered
	} else {
		m := raw.(map[string]any)
		m["jobs"] = filtered
		out = m
	}
	b, err = json.MarshalIndent(out, "", "  ")
	if err != nil {
		return false, err
	}
	return true, os.WriteFile(cronFile, append(b, '\n'), 0644)
}

func ToggleHook(item data.OcHook) (bool, error) {
	return patchJSON(openclawJSON, func(cfg map[string]any) bool {
		hooks, ok := cfg["hooks"].(map[string]any)
		if !ok {
			return false
		}
		internal, ok := hooks["internal"].(map[string]any)
		if !ok {
			return false
		}
		entries, ok := internal["entries"].(map[string]any)
		if !ok {
			return false
		}
		entry, ok := entries[item.ItemName()].(map[string]any)
		if !ok {
			return false
		}
		entry["enabled"] = !item.Enabled
		return true
	})
}

func ToggleCron(item data.OcCronJob) (bool, error) {
	b, err := os.ReadFile(cronFile)
	if err != nil {
		return false, err
	}
	var raw any
	if err := json.Unmarshal(b, &raw); err != nil {
		return false, err
	}

	isArray := false
	var jobs []map[string]any
	switch v := raw.(type) {
	case []any:
		isArray = true
		for _, j := range v {
			if m, ok := j.(map[string]any); ok {
				jobs = append(jobs, m)
			}
		}
	case map[string]any:
		if js, ok := v["jobs"].([]any); ok {
			for _, j := range js {
				if m, ok := j.(map[string]any); ok {
					jobs = append(jobs, m)
				}
			}
		}
	}

	found := false
	for _, j := range jobs {
		if j["id"] == item.ID || j["name"] == item.ItemName() {
			j["enabled"] = !item.Enabled
			found = true
		}
	}
	if !found {
		return false, nil
	}

	var out any
	if isArray {
		out = jobs
	} else {
		m := raw.(map[string]any)
		m["jobs"] = jobs
		out = m
	}
	b, err = json.MarshalIndent(out, "", "  ")
	if err != nil {
		return false, err
	}
	return true, os.WriteFile(cronFile, append(b, '\n'), 0644)
}

func ToggleWebhook(item data.OcWebhook) (bool, error) {
	return patchJSON(openclawJSON, func(cfg map[string]any) bool {
		plugins, ok := cfg["plugins"].(map[string]any)
		if !ok {
			return false
		}
		entries, ok := plugins["entries"].(map[string]any)
		if !ok {
			return false
		}
		webhooks, ok := entries["webhooks"].(map[string]any)
		if !ok {
			return false
		}
		routes, ok := webhooks["routes"].(map[string]any)
		if !ok {
			return false
		}
		route, ok := routes[item.ItemName()].(map[string]any)
		if !ok {
			return false
		}
		route["enabled"] = !item.Enabled
		return true
	})
}

func patchJSON(path string, mutate func(map[string]any) bool) (bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	var cfg map[string]any
	if err := json.Unmarshal(b, &cfg); err != nil {
		return false, err
	}
	if !mutate(cfg) {
		return false, nil
	}
	b, err = json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return false, err
	}
	return true, os.WriteFile(path, append(b, '\n'), 0644)
}
