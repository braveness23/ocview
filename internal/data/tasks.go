package data

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func LoadTasks() []OcTaskRun {
	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".openclaw", "tasks", "runs.sqlite")

	if _, err := os.Stat(dbPath); err != nil {
		return nil
	}

	db, err := sql.Open("sqlite", "file:"+dbPath+"?mode=ro")
	if err != nil {
		return nil
	}
	defer db.Close()

	rows, err := db.Query(
		`SELECT task_id, runtime, label, status, terminal_summary, created_at, ended_at, source_id, error
		 FROM task_runs ORDER BY created_at DESC LIMIT 200`,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var items []OcTaskRun
	for rows.Next() {
		var (
			id, runtime, status    string
			label, summary         sql.NullString
			sourceID, errorMsg     sql.NullString
			createdAt, endedAt     sql.NullInt64
		)
		if err := rows.Scan(&id, &runtime, &label, &status, &summary, &createdAt, &endedAt, &sourceID, &errorMsg); err != nil {
			continue
		}
		name := label.String
		if name == "" && len(id) >= 8 {
			name = id[:8]
		}
		items = append(items, OcTaskRun{
			ID:        id,
			Name_:     name,
			Runtime:   runtime,
			Label:     label.String,
			Status:    status,
			Summary:   summary.String,
			CreatedAt: createdAt.Int64,
			EndedAt:   endedAt.Int64,
			SourceID:  sourceID.String,
			ErrorMsg:  errorMsg.String,
		})
	}
	return items
}
