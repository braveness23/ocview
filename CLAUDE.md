# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
bun run dev          # Run directly from source (no build step)
bun run build        # Bundle to dist/index.js via build.ts
bun test             # Run all tests
bun test tests/data.test.ts   # Run a single test file
```

## Project direction

ocview started as a read-only browser for OpenClaw state. The goal is to make it a **can't-live-without diagnostic and control tool** — the first thing you open when something feels off with Sparky, and a place to understand what's actually running under the hood.

Guiding principle: every item should be fully readable, not just listed. If a user can see that a hook exists but can't see what it does or where its code lives, the tool has failed.

**Completed:**
- Full content viewing with scroll (skills, workspace files, memory chunks, hooks)
- Session transcript viewer with collapsible tool calls
- Live service status header (running/stopped/failed, since, socket health, version)
- Actions: `o` open in `$EDITOR` (auto-reloads on exit), `r` reload data, `t` toggle hook/cron enabled
- Session channel extraction from filename prefix or JSONL metadata fields

## Architecture

**ocview** is a terminal UI (TUI) for browsing OpenClaw state — built with [Ink](https://github.com/vadimdemedes/ink) (React for terminals) and Bun.

### Data flow

`src/index.tsx` → `loadAll()` → all loaders run in parallel → `<App data={...} />`

All data is loaded once at startup (no live refresh). Each category has a dedicated loader in `src/data/`:

| Loader | Source |
|--------|--------|
| `skills.ts` | `~/.openclaw/skills/` (installed) + npm global `openclaw/skills/` (built-in) |
| `hooks.ts` | `~/.openclaw/openclaw.json` → `hooks.internal.entries` |
| `models.ts` | `~/.openclaw/openclaw.json` → `models.providers.litellm.models` |
| `workspace.ts` | `~/.openclaw/workspace/` markdown files |
| `mcp.ts` | `~/.openclaw/openclaw.json` → `mcp.servers` |
| `sessions.ts` | `~/.openclaw/agents/main/sessions/*.jsonl` |
| `cron.ts` | `~/.openclaw/cron/jobs.json` |
| `memory.ts` | `~/.openclaw/memory/main.sqlite` |

### Component structure

```
App (app.tsx)           — state: active panel, selection, search, scope, modal, transcriptSession
└── Layout              — two-column layout, delegates to panels
    ├── CategoryPanel   — left panel, category list
    ├── ItemPanel       — right panel, item list + search bar
    └── StatusBar       — bottom key-binding hints (context-aware: shows "transcript" for sessions)
DetailModal             — scrollable overlay for all non-session items (q/Esc to close)
TranscriptView          — full-screen session reader; replaces DetailModal for session items
```

`App` owns all state. `useInput` in `App` short-circuits when `transcriptSession` or `modalItem` is set so those views handle their own input without conflict.

### Actions (`src/utils/actions.ts`)

- `openInEditor(filePath)` — exits raw mode, spawns `$EDITOR`, restores raw mode on exit. App auto-reloads after.
- `toggleHook(item)` — rewrites `openclaw.json` flipping `hooks.internal.entries.<name>.enabled`
- `toggleCron(item)` — rewrites `cron/jobs.json` flipping the matching job's `enabled`
- `getEditableFilePath(item)` — returns the file path for skill/workspace/memory items; null for others

Keys: `o` (edit), `t` (toggle), `r` (reload). StatusBar shows `o`/`t` only when relevant to the selected item.

### Service status (`src/data/status.ts`)

Loaded async on mount (systemctl calls take ~200ms). Checks:
1. `systemctl --user is-active` for running/stopped/failed
2. `systemctl --user show --property=ActiveEnterTimestamp` for start time
3. `~/.openclaw/logs/config-health.json` for socket health, then journal as fallback
4. npm global `openclaw/package.json` for version (faster than `openclaw --version`)

Status is displayed in the header between the title bar and panels. Notification messages (reload confirmation, toggle result, errors) replace the status line for 3 seconds.

### Reload flow

App owns `data` and `status` as state (initialized from `initialData` prop). Pressing `r` sets `reloading: true`, which triggers a `useEffect` that calls `loadAll()` + `loadStatus()` synchronously after a 50ms defer (so the "reloading…" spinner renders first).

### View routing (App.tsx)

Session items bypass `DetailModal` entirely — pressing Enter sets `transcriptSession` and renders `TranscriptView` full-screen. All other items open `DetailModal`. Both views call their `onClose` prop to return to the main layout.

### DetailModal scroll

Scrollable kinds: `skill`, `workspace`, `memory`. For these, content is pre-rendered into a flat `string[]` via `wrapText()`, then sliced by `scrollOffset`. Keys: `j/k` (line), `d/u` (half-page), `g/G` (top/bottom). A line counter appears in the footer when content overflows.

Non-scrollable kinds (`hook`, `model`, `mcp`, `cron`) render as short JSX field lists — no scroll state needed.

### TranscriptView

Parses the session's JSONL file on mount via `src/data/transcript.ts`. Builds a flat `DisplayLine[]` from the `Turn[]` array — each turn contributes one or more display lines depending on content length and expand state. Tool calls and tool results render as `▶/▼ [tool] name` headers; pressing Enter on the cursor line toggles expansion inline. Scroll is line-based with a visible cursor (cyan highlight).

### Types

`src/types.ts` is the single source of truth for all data shapes. `AnyItem` is a discriminated union on `kind`. Adding a new category requires: a new `kind` in `CategoryKind`, a new interface, extending `AppData` and `AnyItem`, a new loader, and a new case in `getItemsForCategory`.

Notable fields added for rich detail views:
- `OcSkill.fullContent` — raw SKILL.md text
- `OcWorkspaceFile.fullContent` — full file text (preview is still stored for list view)
- `OcHook.rawConfig` — the raw JSON entry from openclaw.json (hooks are npm-package-implemented; config only controls enabled state)

### Build

`build.ts` bundles with Bun's bundler targeting `bun`, stubs out `react-devtools-core`, and patches the shebang to use the absolute path to `bun` (so the installed binary works without `bun` on `PATH`).

### Skill parsing

Skills support two formats: YAML frontmatter (`---\nname:\ndescription:\n---`) or plain markdown (first `#` heading = name, first non-heading paragraph = description). `parseSkillMd` in `skills.ts` returns `{ name, description, raw }` — `raw` is stored as `fullContent` for the detail view.
