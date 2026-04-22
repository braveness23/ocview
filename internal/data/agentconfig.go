package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func LoadAgentConfig() []OcConfigSection {
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

	var sections []OcConfigSection

	if cfg.Agents != nil && cfg.Agents.Defaults != nil {
		sections = append(sections, buildAgentSection(cfg.Agents.Defaults))
	}
	if cfg.Gateway != nil {
		sections = append(sections, buildGatewaySection(cfg.Gateway))
	}
	if cfg.Tools != nil || cfg.Commands != nil {
		sections = append(sections, buildToolsSection(cfg.Tools, cfg.Commands))
	}

	return sections
}

func buildAgentSection(d *agentDefaults) OcConfigSection {
	var lines []string
	lines = append(lines, "primary model:    "+d.Model.Primary)
	if d.TimeoutSeconds > 0 {
		lines = append(lines, fmt.Sprintf("timeout:          %ds", d.TimeoutSeconds))
	}
	if d.MaxConcurrent > 0 {
		lines = append(lines, fmt.Sprintf("max concurrent:   %d", d.MaxConcurrent))
	}
	if d.Heartbeat != nil && d.Heartbeat.Every != "" {
		lines = append(lines, "heartbeat every:  "+d.Heartbeat.Every)
	}
	if d.ContextPruning != nil {
		lines = append(lines, "", "── Context Pruning", "")
		lines = append(lines, "  mode:  "+d.ContextPruning.Mode)
		if d.ContextPruning.TTL != "" {
			lines = append(lines, "  ttl:   "+d.ContextPruning.TTL)
		}
	}
	if d.Compaction != nil {
		lines = append(lines, "", "── Compaction", "")
		lines = append(lines, "  mode:  "+d.Compaction.Mode)
	}
	if d.MemorySearch != nil {
		enabled := "no"
		if d.MemorySearch.Enabled {
			enabled = "yes"
		}
		lines = append(lines, "", "── Memory Search", "")
		lines = append(lines, "  enabled:   "+enabled)
		if d.MemorySearch.Model != "" {
			lines = append(lines, "  model:     "+d.MemorySearch.Model)
		}
		if d.MemorySearch.Remote != nil {
			lines = append(lines, "  endpoint:  "+d.MemorySearch.Remote.BaseURL)
		}
	}
	if d.Subagents != nil {
		lines = append(lines, "", "── Subagents", "")
		lines = append(lines, fmt.Sprintf("  max concurrent:  %d", d.Subagents.MaxConcurrent))
	}
	return OcConfigSection{
		ID:      "config#agent-defaults",
		Name_:   "Agent Defaults",
		Summary: d.Model.Primary,
		Lines:   lines,
	}
}

func buildGatewaySection(gw *gatewaySection) OcConfigSection {
	var lines []string
	lines = append(lines, fmt.Sprintf("port:   %d", gw.Port))
	lines = append(lines, "mode:   "+gw.Mode)
	lines = append(lines, "bind:   "+gw.Bind)
	if gw.Auth != nil {
		lines = append(lines, "auth:   "+gw.Auth.Mode)
	}
	if gw.Tailscale != nil {
		lines = append(lines, "", "── Tailscale", "")
		lines = append(lines, "  mode:          "+gw.Tailscale.Mode)
	}
	if gw.Nodes != nil && len(gw.Nodes.DenyCommands) > 0 {
		lines = append(lines, "", "── Denied node commands", "")
		for _, cmd := range gw.Nodes.DenyCommands {
			lines = append(lines, "  • "+cmd)
		}
	}
	summary := fmt.Sprintf("port %d  bind=%s  mode=%s", gw.Port, gw.Bind, gw.Mode)
	return OcConfigSection{
		ID:      "config#gateway",
		Name_:   "Gateway",
		Summary: summary,
		Lines:   lines,
	}
}

func buildToolsSection(tools *toolsSection, cmds *commandsSection) OcConfigSection {
	var lines []string
	var summaryParts []string

	if tools != nil {
		lines = append(lines, "── Tools", "")
		lines = append(lines, "  profile:            "+tools.Profile)
		summaryParts = append(summaryParts, "profile="+tools.Profile)
		if tools.Exec != nil {
			lines = append(lines, "  exec host:          "+tools.Exec.Host)
			lines = append(lines, "  exec security:      "+tools.Exec.Security)
		}
		if tools.FS != nil {
			wsOnly := "no"
			if tools.FS.WorkspaceOnly {
				wsOnly = "yes"
			}
			lines = append(lines, "  fs workspace-only:  "+wsOnly)
		}
	}

	if cmds != nil {
		lines = append(lines, "", "── Commands", "")
		yesno := func(v bool) string {
			if v {
				return "enabled"
			}
			return "disabled"
		}
		lines = append(lines, "  bash:          "+yesno(cmds.Bash))
		lines = append(lines, "  config:        "+yesno(cmds.Config))
		lines = append(lines, "  mcp:           "+yesno(cmds.MCP))
		lines = append(lines, "  plugins:       "+yesno(cmds.Plugins))
		lines = append(lines, "  debug:         "+yesno(cmds.Debug))
		lines = append(lines, "  restart:       "+yesno(cmds.Restart))
		if cmds.Native != "" {
			lines = append(lines, "  native:        "+cmds.Native)
		}
		if cmds.NativeSkills != "" {
			lines = append(lines, "  native skills: "+cmds.NativeSkills)
		}
		if cmds.OwnerDisplay != "" {
			lines = append(lines, "  owner display: "+cmds.OwnerDisplay)
		}
	}

	return OcConfigSection{
		ID:      "config#tools",
		Name_:   "Tools & Commands",
		Summary: strings.Join(summaryParts, "  "),
		Lines:   lines,
	}
}
