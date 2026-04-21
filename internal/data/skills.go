package data

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var fmRe = regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---`)
var fmNameRe = regexp.MustCompile(`(?m)^name:\s*(.+)$`)
var fmDescRe = regexp.MustCompile(`(?m)^description:\s*(.+)$`)

func parseSkillMd(filePath string) (name, description, raw string) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return "", "", ""
	}
	raw = string(b)

	if m := fmRe.FindStringSubmatch(raw); m != nil {
		block := m[1]
		if nm := fmNameRe.FindStringSubmatch(block); nm != nil {
			name = strings.TrimSpace(nm[1])
		}
		if dm := fmDescRe.FindStringSubmatch(block); dm != nil {
			description = strings.TrimSpace(dm[1])
		}
		if name != "" || description != "" {
			return
		}
	}

	pastHeading := false
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if name == "" && strings.HasPrefix(trimmed, "# ") {
			name = strings.TrimLeft(trimmed, "# ")
			pastHeading = true
			continue
		}
		if pastHeading && description == "" && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			description = trimmed
			break
		}
	}
	return
}

func findBuiltInSkillsDir() string {
	if out, err := exec.Command("npm", "root", "-g").Output(); err == nil {
		p := filepath.Join(strings.TrimSpace(string(out)), "openclaw", "skills")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, ".npm-global", "lib", "node_modules", "openclaw", "skills"),
		"/usr/local/lib/node_modules/openclaw/skills",
		"/usr/lib/node_modules/openclaw/skills",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func loadSkillsFromDir(dir, scope string) []OcSkill {
	if _, err := os.Stat(dir); err != nil {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var items []OcSkill
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillFile := filepath.Join(dir, entry.Name(), "SKILL.md")
		if _, err := os.Stat(skillFile); err != nil {
			continue
		}
		name, desc, raw := parseSkillMd(skillFile)
		if name == "" {
			name = entry.Name()
		}
		items = append(items, OcSkill{
			ID:          scope + "#" + entry.Name(),
			Name_:       name,
			Description: desc,
			Scope:       scope,
			FilePath:    skillFile,
			FullContent: raw,
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name_ < items[j].Name_ })
	return items
}

func LoadSkills() []OcSkill {
	var result []OcSkill
	if dir := findBuiltInSkillsDir(); dir != "" {
		result = append(result, loadSkillsFromDir(dir, "built-in")...)
	}
	home, _ := os.UserHomeDir()
	result = append(result, loadSkillsFromDir(filepath.Join(home, ".openclaw", "skills"), "installed")...)
	return result
}
