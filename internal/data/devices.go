package data

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type deviceEntry struct {
	DeviceID       string   `json:"deviceId"`
	Platform       string   `json:"platform"`
	ClientID       string   `json:"clientId"`
	ClientMode     string   `json:"clientMode"`
	Role           string   `json:"role"`
	ApprovedScopes []string `json:"approvedScopes"`
	DisplayName    string   `json:"displayName"`
	CreatedAtMs    int64    `json:"createdAtMs"`
}

func LoadDevices() []OcDevice {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".openclaw", "devices")

	var items []OcDevice
	items = append(items, loadDeviceFile(filepath.Join(dir, "paired.json"), "paired")...)
	items = append(items, loadDeviceFile(filepath.Join(dir, "pending.json"), "pending")...)
	return items
}

func loadDeviceFile(path, status string) []OcDevice {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var file map[string]deviceEntry
	if err := json.Unmarshal(b, &file); err != nil {
		return nil
	}
	var out []OcDevice
	for _, d := range file {
		name := d.DisplayName
		if name == "" {
			name = d.ClientID
		}
		if name == "" && len(d.DeviceID) >= 8 {
			name = d.DeviceID[:8]
		}
		out = append(out, OcDevice{
			ID:        d.DeviceID,
			Name_:     name,
			DeviceID:  d.DeviceID,
			Platform:  d.Platform,
			Role:      d.Role,
			Scopes:    d.ApprovedScopes,
			ClientID:  d.ClientID,
			CreatedAt: d.CreatedAtMs,
			Status:    status,
		})
	}
	return out
}
