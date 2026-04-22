# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go build -o ocview .          # Build binary
go run .                      # Run directly from source
ln -sf $(pwd)/ocview ~/.local/bin/ocview # Symlink into PATH (one-time setup; go build updates it automatically)
```

## Project direction

ocview is a terminal UI (TUI) for browsing OpenClaw state — the first thing you open when something feels off with Sparky, and a place to understand what's actually running under the hood.

Guiding principle: every item should be fully readable, not just listed. If a user can see that a hook exists but can't see what it does or where its code lives, the tool has failed.

**Planned:**
- Log viewer category: tail/browse `journalctl --user -u openclaw-gateway.service` output and files in `~/.openclaw/logs/` (config-health.json, sync-token.log, config-audit.jsonl). Should support live-tail mode and filtering by level/keyword.

**Completed:**
- Full content viewing with scroll (skills, workspace files, memory chunks, hooks)
- Session transcript viewer with collapsible tool calls
- Live service status header (running/stopped/failed, since, socket health, version)
- Actions: `o` open in `$EDITOR` (auto-reloads on exit), `r` reload data, `t` toggle hook/cron enabled
- Session channel extraction from filename prefix or JSONL metadata fields
- Skill management: `n` create new installed skill (prompts for dir name, scaffolds SKILL.md, opens in editor), `d` delete installed skill (with y/n confirmation); built-in skills are protected
- Cron job deletion: `d` removes the job from `jobs.json` (with confirmation)
- Config audit log category

## Architecture

**ocview** is built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea) (Charm TUI framework) with [Lip Gloss](https://github.com/charmbracelet/lipgloss) for styling.

### Data flow

`main.go` → `loadAll()` → all loaders run in parallel → initial `model` → Bubble Tea `tea.NewProgram`

All data is loaded once at startup (no live refresh unless `r` is pressed). Each category has a dedicated loader in `internal/data/`:

| Loader | Source |
|--------|--------|
| `skills.go` | `~/.openclaw/skills/` (installed) + npm global `openclaw/skills/` (built-in) |
| `hooks.go` | `~/.openclaw/openclaw.json` → `hooks.internal.entries` |
| `models.go` | `~/.openclaw/openclaw.json` → `models.providers.litellm.models` |
| `workspace.go` | `~/.openclaw/workspace/` markdown files |
| `mcp.go` | `~/.openclaw/openclaw.json` → `mcp.servers` |
| `sessions.go` | `~/.openclaw/agents/main/sessions/*.jsonl` |
| `cron.go` | `~/.openclaw/cron/jobs.json` |
| `memory.go` | `~/.openclaw/memory/main.sqlite` |
| `audit.go` | `~/.openclaw/logs/config-audit.jsonl` |
| `webhooks.go` | `~/.openclaw/openclaw.json` → `plugins.entries.webhooks` |
| `status.go` | `systemctl`, `config-health.json`, npm package.json |

### Component structure

Bubble Tea model (`internal/ui/model.go`) owns all state. Views:
- **Category panel** — left column, category list
- **Item panel** — right column, filtered item list + search bar
- **Detail view** — scrollable overlay for non-session items
- **Transcript view** — full-screen session reader with collapsible tool calls
- **Status bar** — bottom key-binding hints (context-aware)

### Actions

- `o` — open item's file in `$EDITOR` (suspends Bubble Tea, restores + reloads on exit)
- `t` — toggle enabled (hooks, cron, webhooks) — rewrites the backing JSON in place
- `r` — reload all data
- `n` — new installed skill (prompts for dir name, scaffolds `SKILL.md`, opens in editor)
- `d` — delete (installed skills, cron jobs) — shows `y/n` confirmation in status bar

#### Input mode priority
`transcriptView` → `detailView` → `newSkillPrompt` → `confirmDelete` → `searchActive` → normal navigation. Each mode short-circuits the ones below it.

### Service status

Loaded async on startup (~200ms systemctl calls). Checks:
1. `systemctl --user is-active` for running/stopped/failed
2. `systemctl --user show --property=ActiveEnterTimestamp` for start time
3. `~/.openclaw/logs/config-health.json` for socket health
4. npm global `openclaw/package.json` for version

### Cron job format

`~/.openclaw/cron/jobs.json` uses a format where `schedule` is a nested object `{expr, tz, kind}` and the prompt text is in `payload.text`. The loader normalises this into a flat shape. Don't assume a top-level `schedule` string or `command` field.

### Skill parsing

Skills support two formats: YAML frontmatter (`---\nname:\ndescription:\n---`) or plain markdown (first `#` heading = name, first non-heading paragraph = description).
