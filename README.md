# ocview

Terminal UI for browsing [OpenClaw](https://github.com/openclaw/openclaw) state — skills, hooks, models, sessions, cron jobs, memory, webhooks, and more.

![](images/view.png)

## Build

Requires Go 1.22+.

```bash
git clone https://github.com/braveness23/ocview
cd ocview
go build -o ocview-go .
```

Or just run directly:

```bash
go run .
```

## Usage

```bash
./ocview-go
```

Reads live state from `~/.openclaw/`. No arguments needed.

## Key Bindings

### Main view

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate |
| `→` / `Enter` / `Tab` | Open items / open detail |
| `←` / `Esc` | Back |
| `/` | Search / filter |
| `s` | Cycle scope filter (Skills category) |
| `n` | New skill (Skills category) |
| `o` | Open in `$EDITOR` |
| `t` | Toggle enabled (hooks, cron, webhooks) |
| `d` | Delete (installed skills, cron jobs) |
| `r` | Reload all data |
| `q` / `Ctrl+C` | Quit |

### Detail / Transcript

| Key | Action |
|-----|--------|
| `↑` / `↓` | Scroll line |
| `PgDn` / `PgUp` | Scroll half-page |
| `Home` / `End` | Top / bottom |
| `Enter` | Expand/collapse tool call (transcript) |
| `o` | Open in `$EDITOR` |
| `q` / `Esc` | Close |

Vim aliases `j/k`, `g/G`, `d/u` also work everywhere.

## Categories

| Category | Source |
|----------|--------|
| Skills | `~/.openclaw/skills/` + npm global built-ins |
| Hooks | `~/.openclaw/openclaw.json` → `hooks.internal` |
| Models | `~/.openclaw/openclaw.json` → `models.providers` |
| Workspace | `~/.openclaw/workspace/*.md` |
| MCP | `~/.openclaw/openclaw.json` → `mcp.servers` |
| Sessions | `~/.openclaw/agents/main/sessions/*.jsonl` |
| Cron | `~/.openclaw/cron/jobs.json` |
| Memory | `~/.openclaw/memory/main.sqlite` |
| Updates | `CHANGELOG.md` + `update-check.json` |
| Webhooks | `~/.openclaw/openclaw.json` → `plugins.entries.webhooks` |
| Audit Log | `~/.openclaw/logs/config-audit.jsonl` |
