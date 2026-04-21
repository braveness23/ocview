package data

// Shared JSON structures for parsing openclaw.json.

type openclawJSON struct {
	Hooks   *hooksSection   `json:"hooks"`
	Models  *modelsSection  `json:"models"`
	MCP     *mcpSection     `json:"mcp"`
	Plugins *pluginsSection `json:"plugins"`
}

type hooksSection struct {
	Internal *hooksInternal `json:"internal"`
}

type hooksInternal struct {
	Enabled bool                       `json:"enabled"`
	Entries map[string]map[string]any  `json:"entries"`
}

type modelsSection struct {
	Providers map[string]modelProvider `json:"providers"`
}

type modelProvider struct {
	Models []modelEntry `json:"models"`
}

type modelEntry struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Reasoning     bool        `json:"reasoning"`
	ContextWindow int         `json:"contextWindow"`
	MaxTokens     int         `json:"maxTokens"`
	Cost          *modelCost  `json:"cost"`
}

type modelCost struct {
	Input  float64 `json:"input"`
	Output float64 `json:"output"`
}

type mcpSection struct {
	Servers map[string]mcpServerEntry `json:"servers"`
}

type mcpServerEntry struct {
	URL       string            `json:"url"`
	Command   string            `json:"command"`
	Args      []string          `json:"args"`
	Transport string            `json:"transport"`
	Headers   map[string]string `json:"headers"`
	Enabled   *bool             `json:"enabled"`
}

type pluginsSection struct {
	Entries *pluginEntries `json:"entries"`
}

type pluginEntries struct {
	Webhooks *webhooksPlugin `json:"webhooks"`
}

type webhooksPlugin struct {
	Routes map[string]webhookRoute `json:"routes"`
}

type webhookRoute struct {
	Enabled      *bool      `json:"enabled"`
	Path         string     `json:"path"`
	SessionKey   string     `json:"sessionKey"`
	Secret       any        `json:"secret"` // string or SecretRef
	ControllerID string     `json:"controllerId"`
	Description  string     `json:"description"`
}
