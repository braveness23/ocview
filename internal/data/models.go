package data

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func LoadModels() []OcModel {
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
	if cfg.Models == nil {
		return nil
	}

	var items []OcModel
	for providerName, provider := range cfg.Models.Providers {
		for _, m := range provider.Models {
			name := m.Name
			if name == "" {
				name = m.ID
			}
			var costIn, costOut float64
			if m.Cost != nil {
				costIn = m.Cost.Input
				costOut = m.Cost.Output
			}
			items = append(items, OcModel{
				ID:            providerName + "/" + m.ID,
				Name_:         name,
				Provider:      providerName,
				Reasoning:     m.Reasoning,
				ContextWindow: m.ContextWindow,
				MaxTokens:     m.MaxTokens,
				CostInput:     costIn,
				CostOutput:    costOut,
			})
		}
	}
	return items
}
