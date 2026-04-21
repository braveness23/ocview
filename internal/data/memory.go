package data

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func LoadMemory() []OcMemoryChunk {
	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".openclaw", "memory", "main.sqlite")

	if _, err := os.Stat(dbPath); err != nil {
		return nil
	}

	db, err := sql.Open("sqlite", "file:"+dbPath+"?mode=ro")
	if err != nil {
		return nil
	}
	defer db.Close()

	rows, err := db.Query(
		`SELECT id, path, source, start_line, end_line, model, text, updated_at
		 FROM chunks ORDER BY updated_at DESC`,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var items []OcMemoryChunk
	for rows.Next() {
		var r struct {
			id        string
			path      string
			source    string
			startLine int
			endLine   int
			model     string
			text      string
			updatedAt int64
		}
		if err := rows.Scan(&r.id, &r.path, &r.source, &r.startLine, &r.endLine, &r.model, &r.text, &r.updatedAt); err != nil {
			continue
		}
		name := fmt.Sprintf("%s:%d-%d", filepath.Base(r.path), r.startLine, r.endLine)
		items = append(items, OcMemoryChunk{
			ID:        r.id,
			Name_:     name,
			Path:      r.path,
			Source:    r.source,
			StartLine: r.startLine,
			EndLine:   r.endLine,
			Model:     r.model,
			Text:      r.text,
			UpdatedAt: r.updatedAt,
		})
	}
	return items
}
