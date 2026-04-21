package data

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func checkCommand(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func checkHTTPURL(url string) bool {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Head(url)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode > 0
}

func LoadMCP() []OcMcpServer {
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
	if cfg.MCP == nil {
		return nil
	}

	var items []OcMcpServer
	for name, entry := range cfg.MCP.Servers {
		transport := entry.Transport
		if transport == "" {
			if entry.Command != "" {
				transport = "stdio"
			} else {
				transport = "unknown"
			}
		}
		enabled := true
		if entry.Enabled != nil {
			enabled = *entry.Enabled
		}

		var deps []McpDependency
		if entry.Command != "" {
			deps = append(deps, McpDependency{Name: entry.Command, Met: checkCommand(entry.Command)})
		} else if entry.URL != "" {
			host := entry.URL
			// Extract just the host for display
			if idx := strings.Index(host, "://"); idx >= 0 {
				host = host[idx+3:]
				if idx2 := strings.IndexByte(host, '/'); idx2 >= 0 {
					host = host[:idx2]
				}
			}
			deps = append(deps, McpDependency{Name: host, Met: checkHTTPURL(entry.URL)})
		}

		available := enabled
		for _, d := range deps {
			if !d.Met {
				available = false
				break
			}
		}

		items = append(items, OcMcpServer{
			ID:           "mcp#" + name,
			Name_:        name,
			URL:          entry.URL,
			Command:      entry.Command,
			Args:         entry.Args,
			Transport:    transport,
			Headers:      entry.Headers,
			Enabled:      enabled,
			Available:    available,
			Dependencies: deps,
		})
	}
	return items
}
