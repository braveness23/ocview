package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func describeSecret(raw any) string {
	if raw == nil {
		return "(none)"
	}
	if s, ok := raw.(string); ok {
		if s == "" {
			return "(none)"
		}
		return "••••••••"
	}
	if m, ok := raw.(map[string]any); ok {
		source, _ := m["source"].(string)
		provider, _ := m["provider"].(string)
		id, _ := m["id"].(string)
		return fmt.Sprintf("%s:%s/%s", source, provider, id)
	}
	return "(unknown)"
}

func LoadWebhooks() []OcWebhook {
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
	if cfg.Plugins == nil || cfg.Plugins.Entries == nil ||
		cfg.Plugins.Entries.Webhooks == nil {
		return nil
	}

	var items []OcWebhook
	for routeID, route := range cfg.Plugins.Entries.Webhooks.Routes {
		enabled := true
		if route.Enabled != nil {
			enabled = *route.Enabled
		}
		path := route.Path
		if path == "" {
			path = "/plugins/webhooks/" + routeID
		}
		controllerID := route.ControllerID
		if controllerID == "" {
			controllerID = "webhooks/" + routeID
		}
		items = append(items, OcWebhook{
			ID:           "webhook#" + routeID,
			Name_:        routeID,
			Enabled:      enabled,
			Path:         path,
			SessionKey:   route.SessionKey,
			Secret:       describeSecret(route.Secret),
			ControllerID: controllerID,
			Description:  route.Description,
		})
	}
	return items
}
