package data

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func LoadCron() []OcCronJob {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".openclaw", "cron", "jobs.json")

	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	// Try array first, then {jobs: [...]} wrapper
	var rawJobs []json.RawMessage
	if err := json.Unmarshal(b, &rawJobs); err != nil {
		var wrapper struct {
			Jobs []json.RawMessage `json:"jobs"`
		}
		if err2 := json.Unmarshal(b, &wrapper); err2 != nil {
			return nil
		}
		rawJobs = wrapper.Jobs
	}

	var items []OcCronJob
	for i, raw := range rawJobs {
		var job struct {
			ID          string          `json:"id"`
			Name        string          `json:"name"`
			Schedule    json.RawMessage `json:"schedule"`
			Command     string          `json:"command"`
			Payload     *struct {
				Text    string `json:"text"`
				Message string `json:"message"`
			} `json:"payload"`
			Enabled     *bool  `json:"enabled"`
			Description string `json:"description"`
		}
		if err := json.Unmarshal(raw, &job); err != nil {
			continue
		}

		// Parse schedule (string or {expr, tz} object)
		var schedStr, tz string
		var schedObj struct {
			Expr string `json:"expr"`
			TZ   string `json:"tz"`
		}
		if err := json.Unmarshal(job.Schedule, &schedStr); err != nil {
			if err2 := json.Unmarshal(job.Schedule, &schedObj); err2 == nil {
				schedStr = schedObj.Expr
				tz = schedObj.TZ
			}
		}
		if tz != "" {
			schedStr = schedStr + " (" + tz + ")"
		}

		command := job.Command
		if command == "" && job.Payload != nil {
			command = job.Payload.Text
			if command == "" {
				command = job.Payload.Message
			}
		}

		id := job.ID
		if id == "" {
			id = fmt.Sprintf("cron#%d", i)
		}
		name := job.Name
		if name == "" {
			name = id
		}

		enabled := true
		if job.Enabled != nil {
			enabled = *job.Enabled
		}

		items = append(items, OcCronJob{
			ID:          id,
			Name_:       name,
			Schedule:    schedStr,
			Command:     command,
			Enabled:     enabled,
			Description: job.Description,
			LastRuns:    loadCronRuns(home, id),
		})
	}
	return items
}

func loadCronRuns(home, jobID string) []CronRunRecord {
	if jobID == "" {
		return nil
	}
	path := filepath.Join(home, ".openclaw", "cron", "runs", jobID+".jsonl")
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var all []CronRunRecord
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var row struct {
			Ts         int64  `json:"ts"`
			Status     string `json:"status"`
			Summary    string `json:"summary"`
			DurationMs int64  `json:"durationMs"`
			NextRunMs  int64  `json:"nextRunAtMs"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
			continue
		}
		all = append(all, CronRunRecord{
			Ts:         row.Ts,
			Status:     row.Status,
			Summary:    row.Summary,
			DurationMs: row.DurationMs,
			NextRunMs:  row.NextRunMs,
		})
	}
	// Return newest-first (last N entries reversed)
	const maxRuns = 10
	if len(all) > maxRuns {
		all = all[len(all)-maxRuns:]
	}
	for i, j := 0, len(all)-1; i < j; i, j = i+1, j-1 {
		all[i], all[j] = all[j], all[i]
	}
	return all
}
