package data

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var workspaceFiles = []string{
	"IDENTITY.md",
	"SOUL.md",
	"USER.md",
	"AGENTS.md",
	"TOOLS.md",
	"MEMORY.md",
	"HEARTBEAT.md",
	"BOOTSTRAP.md",
}

func readWorkspaceFile(filePath, id, name string) *OcWorkspaceFile {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}
	info, err := os.Stat(filePath)
	if err != nil {
		return nil
	}
	content := string(b)
	preview := content
	if len(preview) > 500 {
		preview = preview[:500]
	}
	wordCount := len(strings.Fields(content))
	return &OcWorkspaceFile{
		ID:           id,
		Name_:        name,
		FilePath:     filePath,
		Preview:      preview,
		FullContent:  content,
		WordCount:    wordCount,
		LastModified: info.ModTime().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func LoadWorkspace() []OcWorkspaceFile {
	home, _ := os.UserHomeDir()
	workspace := filepath.Join(home, ".openclaw", "workspace")

	var items []OcWorkspaceFile
	for _, file := range workspaceFiles {
		path := filepath.Join(workspace, file)
		if _, err := os.Stat(path); err != nil {
			continue
		}
		if item := readWorkspaceFile(path, "workspace#"+file, file); item != nil {
			items = append(items, *item)
		}
	}

	memDir := filepath.Join(workspace, "memory")
	if entries, err := os.ReadDir(memDir); err == nil {
		var memFiles []string
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				memFiles = append(memFiles, e.Name())
			}
		}
		sort.Sort(sort.Reverse(sort.StringSlice(memFiles)))
		for _, file := range memFiles {
			path := filepath.Join(memDir, file)
			name := "memory/" + file
			if item := readWorkspaceFile(path, "workspace#"+name, name); item != nil {
				items = append(items, *item)
			}
		}
	}
	return items
}
