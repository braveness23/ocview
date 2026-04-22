package data

// Shared JSON structures for parsing openclaw.json.

type openclawJSON struct {
	Hooks    *hooksSection    `json:"hooks"`
	Models   *modelsSection   `json:"models"`
	MCP      *mcpSection      `json:"mcp"`
	Plugins  *pluginsSection  `json:"plugins"`
	Agents   *agentsSection   `json:"agents"`
	Gateway  *gatewaySection  `json:"gateway"`
	Tools    *toolsSection    `json:"tools"`
	Commands *commandsSection `json:"commands"`
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
	Enabled      *bool  `json:"enabled"`
	Path         string `json:"path"`
	SessionKey   string `json:"sessionKey"`
	Secret       any    `json:"secret"` // string or SecretRef
	ControllerID string `json:"controllerId"`
	Description  string `json:"description"`
}

type agentsSection struct {
	Defaults *agentDefaults `json:"defaults"`
}

type agentDefaults struct {
	Model struct {
		Primary string `json:"primary"`
	} `json:"model"`
	Workspace      string `json:"workspace"`
	TimeoutSeconds int    `json:"timeoutSeconds"`
	MaxConcurrent  int    `json:"maxConcurrent"`
	Heartbeat      *struct {
		Every string `json:"every"`
	} `json:"heartbeat"`
	ContextPruning *struct {
		Mode string `json:"mode"`
		TTL  string `json:"ttl"`
	} `json:"contextPruning"`
	Compaction *struct {
		Mode string `json:"mode"`
	} `json:"compaction"`
	MemorySearch *struct {
		Enabled bool   `json:"enabled"`
		Model   string `json:"model"`
		Remote  *struct {
			BaseURL string `json:"baseUrl"`
		} `json:"remote"`
	} `json:"memorySearch"`
	Subagents *struct {
		MaxConcurrent int `json:"maxConcurrent"`
	} `json:"subagents"`
}

type gatewaySection struct {
	Port int    `json:"port"`
	Mode string `json:"mode"`
	Bind string `json:"bind"`
	Auth *struct {
		Mode string `json:"mode"`
	} `json:"auth"`
	Tailscale *struct {
		Mode        string `json:"mode"`
		ResetOnExit bool   `json:"resetOnExit"`
	} `json:"tailscale"`
	Nodes *struct {
		DenyCommands []string `json:"denyCommands"`
	} `json:"nodes"`
}

type toolsSection struct {
	Profile string `json:"profile"`
	Exec    *struct {
		Host     string `json:"host"`
		Security string `json:"security"`
	} `json:"exec"`
	FS *struct {
		WorkspaceOnly bool `json:"workspaceOnly"`
	} `json:"fs"`
}

type commandsSection struct {
	Native       string `json:"native"`
	NativeSkills string `json:"nativeSkills"`
	Bash         bool   `json:"bash"`
	Config       bool   `json:"config"`
	MCP          bool   `json:"mcp"`
	Plugins      bool   `json:"plugins"`
	Debug        bool   `json:"debug"`
	Restart      bool   `json:"restart"`
	OwnerDisplay string `json:"ownerDisplay"`
}
