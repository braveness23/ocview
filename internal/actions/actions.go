package actions

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	case data.OcCronJob:
		_ = v
		return cronFile
	case data.OcHook:
		_ = v
		return openclawJSON
	case data.OcMcpServer:
		_ = v
		return openclawJSON
	case data.OcWebhook:
		_ = v
		return openclawJSON
	case data.OcModel:
		_ = v
		return openclawJSON
	case data.OcLogFile:
		return v.FilePath // empty string for journalctl — disables edit
	case data.OcConfigSection:
		_ = v
		return openclawJSON
	}
	return ""
}

func GetEditLineNumber(item data.AnyItem) int {
	switch v := item.(type) {
	case data.OcMemoryChunk:
		return v.StartLine
	case data.OcHook:
		return lineOf(openclawJSON, `"`+v.ItemName()+`"`)
	case data.OcMcpServer:
		return lineOf(openclawJSON, `"`+v.ItemName()+`"`)
	case data.OcWebhook:
		return lineOf(openclawJSON, `"`+v.ItemName()+`"`)
	case data.OcModel:
		id := v.ID
		if i := strings.LastIndex(id, "/"); i >= 0 {
			id = id[i+1:]
		}
		return lineOf(openclawJSON, `"id": "`+id+`"`)
	case data.OcCronJob:
		if v.ID != "" {
			if line := lineOf(cronFile, `"`+v.ID+`"`); line > 0 {
				return line
			}
		}
		return lineOf(cronFile, `"`+v.ItemName()+`"`)
	}
	return 0
}

func lineOf(path, search string) int {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	content := string(b)
	idx := strings.Index(content, search)
	if idx < 0 {
		return 0
	}
	return strings.Count(content[:idx], "\n") + 1
}

func EditorCmd(filePath string, line int) *exec.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "nano"
	}
	if line <= 0 {
		return exec.Command(editor, filePath)
	}
	base := filepath.Base(editor)
	switch base {
	case "code", "code-insiders":
		return exec.Command(editor, "--goto", fmt.Sprintf("%s:%d", filePath, line))
	case "hx", "micro":
		return exec.Command(editor, fmt.Sprintf("%s:%d", filePath, line))
	default:
		// vim, nvim, vi, nano, emacs, and most others accept +N
		return exec.Command(editor, fmt.Sprintf("+%d", line), filePath)
	}
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
