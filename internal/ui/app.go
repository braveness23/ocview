package ui

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/braveness23/ocview/internal/actions"
	"github.com/braveness23/ocview/internal/data"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── View modes ───────────────────────────────────────────────────────────────

type viewMode int

const (
	viewMain viewMode = iota
	viewDetail
	viewTranscript
)

type activePanel int

const (
	panelCategories activePanel = iota
	panelItems
)

// ─── Messages ─────────────────────────────────────────────────────────────────

type statusLoadedMsg data.ServiceStatus
type dataReloadedMsg struct {
	appData data.AppData
	status  data.ServiceStatus
}
type notifyClearMsg struct{ seq int }
type editorDoneMsg struct{ err error }
type changelogFetchedMsg struct {
	changes []string
	fixes   []string
}
type changelogErrorMsg struct{}

// ─── Category metadata ────────────────────────────────────────────────────────

var categoryOrder = []string{
	"skills", "hooks", "models", "workspace", "mcp",
	"sessions", "cron", "tasks", "memory", "updates",
	"webhooks", "auditlog", "agentconfig", "devices", "logs",
}

var categoryLabel = map[string]string{
	"skills": "Skills", "hooks": "Hooks", "models": "Models",
	"workspace": "Workspace", "mcp": "MCP", "sessions": "Sessions",
	"cron": "Cron", "tasks": "Tasks", "memory": "Memory", "updates": "Updates",
	"webhooks": "Webhooks", "auditlog": "AuditLog",
	"agentconfig": "Config", "devices": "Devices", "logs": "Logs",
}

var categoryPanelLabel = map[string]string{
	"skills": "SKILLS", "hooks": "HOOKS", "models": "MODELS",
	"workspace": "WORKSPACE", "mcp": "MCP SERVERS", "sessions": "SESSIONS",
	"cron": "CRON JOBS", "tasks": "TASK RUNS", "memory": "MEMORY CHUNKS",
	"updates": "RELEASES", "webhooks": "WEBHOOKS", "auditlog": "CONFIG AUDIT LOG",
	"agentconfig": "AGENT CONFIG", "devices": "DEVICES", "logs": "LOGS",
}

// ─── Model ────────────────────────────────────────────────────────────────────

type Model struct {
	// terminal
	width, height int

	// data
	appData data.AppData
	status  *data.ServiceStatus

	// navigation
	mode      viewMode
	panel     activePanel
	catIdx    int
	itemIdx   int
	scope     string // "all" | "built-in" | "installed"

	// search
	searchActive bool
	searchQuery  string

	// new skill name input
	newSkillActive bool
	newSkillName   string

	// confirm delete
	confirmItem data.AnyItem

	// notification
	notification string
	notifySeq    int

	// reloading
	reloading bool

	// detail modal
	modalItem    data.AnyItem
	modalScroll  int
	fetchState   string // "idle" | "fetching" | "done" | "error"
	fetchedChanges []string
	fetchedFixes   []string

	// transcript
	transcriptSession  *data.OcSession
	transcriptTurns    []data.Turn
	transcriptCursor   int
	transcriptOffset   int
	transcriptExpanded map[int]bool
}

func New() Model {
	return Model{
		width:  80,
		height: 24,
		scope:  "all",
		transcriptExpanded: map[int]bool{},
	}
}

// ─── Init ─────────────────────────────────────────────────────────────────────

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadDataCmd(),
		loadStatusCmd(),
	)
}

// ─── Cmds ─────────────────────────────────────────────────────────────────────

func loadDataCmd() tea.Cmd {
	return func() tea.Msg {
		return dataReloadedMsg{
			appData: data.LoadAll(),
			status:  data.LoadStatus(),
		}
	}
}

func loadStatusCmd() tea.Cmd {
	return func() tea.Msg {
		s := data.LoadStatus()
		return statusLoadedMsg(s)
	}
}

func notifyClearCmd(seq int) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(3 * time.Second)
		return notifyClearMsg{seq: seq}
	}
}

func fetchChangelogCmd(version string) tea.Cmd {
	return func() tea.Msg {
		url := getChangelogURL()
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			return changelogErrorMsg{}
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return changelogErrorMsg{}
		}
		parsed := data.ParseChangelog(string(b))
		entry := parsed[version]
		var changes, fixes []string
		for _, c := range entry.Changes {
			changes = append(changes, data.StripPRs(c))
		}
		for _, f := range entry.Fixes {
			fixes = append(fixes, data.StripPRs(f))
		}
		return changelogFetchedMsg{changes: changes, fixes: fixes}
	}
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.MouseMsg:
		if m.mode == viewMain && !m.searchActive && !m.newSkillActive && m.confirmItem == nil {
			items := m.filteredItems()
			if msg.Type == tea.MouseWheelDown {
				m.itemIdx = clamp(m.itemIdx+1, 0, len(items)-1)
			} else if msg.Type == tea.MouseWheelUp {
				m.itemIdx = clamp(m.itemIdx-1, 0, len(items)-1)
			}
		}
		return m, nil

	case statusLoadedMsg:
		s := data.ServiceStatus(msg)
		m.status = &s
		return m, nil

	case dataReloadedMsg:
		m.appData = msg.appData
		m.status = &msg.status
		m.reloading = false
		return m.setNotify("Reloaded")

	case notifyClearMsg:
		if msg.seq == m.notifySeq {
			m.notification = ""
		}
		return m, nil

	case editorDoneMsg:
		return m, loadDataCmd()

	case changelogFetchedMsg:
		m.fetchedChanges = msg.changes
		m.fetchedFixes = msg.fixes
		m.fetchState = "done"
		return m, nil

	case changelogErrorMsg:
		m.fetchState = "error"
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// ── Transcript view ──────────────────────────────────────────────────────
	if m.mode == viewTranscript {
		lines := buildTranscriptLines(m.transcriptTurns, m.transcriptExpanded, m.width-2)
		total := len(lines)
		contentH := clamp(m.height-4, 4, m.height-4)

		moveCursor := func(next int) {
			m.transcriptCursor = clamp(next, 0, total-1)
			if m.transcriptCursor < m.transcriptOffset {
				m.transcriptOffset = m.transcriptCursor
			} else if m.transcriptCursor >= m.transcriptOffset+contentH {
				m.transcriptOffset = m.transcriptCursor - contentH + 1
			}
		}

		switch key {
		case "q", "esc":
			m.mode = viewMain
			m.transcriptSession = nil
			m.transcriptTurns = nil
			m.transcriptExpanded = map[int]bool{}
		case "j", "down":
			moveCursor(m.transcriptCursor + 1)
		case "k", "up":
			moveCursor(m.transcriptCursor - 1)
		case "pgdown", "d":
			moveCursor(m.transcriptCursor + contentH/2)
		case "pgup", "u":
			moveCursor(m.transcriptCursor - contentH/2)
		case "home", "g":
			moveCursor(0)
		case "end", "G":
			moveCursor(total - 1)
		case "enter":
			if m.transcriptCursor < len(lines) {
				dl := lines[m.transcriptCursor]
				if dl.isToolHeader {
					if m.transcriptExpanded[dl.turnIdx] {
						delete(m.transcriptExpanded, dl.turnIdx)
					} else {
						m.transcriptExpanded[dl.turnIdx] = true
					}
				}
			}
		}
		return m, nil
	}

	// ── Detail modal ─────────────────────────────────────────────────────────
	if m.mode == viewDetail {
		modalWidth := min(m.width, 90)
		textWidth := modalWidth - 6
		chromeRows := 10
		contentH := clamp(m.height-chromeRows, 5, m.height-4)
		lines := renderDetailLines(m.modalItem, textWidth, m.fetchedChanges, m.fetchedFixes, m.fetchState)
		total := len(lines)
		maxOffset := clamp(total-contentH, 0, total)

		switch key {
		case "q", "esc":
			m.mode = viewMain
			m.modalItem = nil
			m.modalScroll = 0
			m.fetchState = "idle"
			m.fetchedChanges = nil
			m.fetchedFixes = nil
		case "o":
			path := actions.GetEditableFilePath(m.modalItem)
			if path != "" {
				line := actions.GetEditLineNumber(m.modalItem)
				m.mode = viewMain
				m.modalItem = nil
				cmd := actions.EditorCmd(path, line)
				return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
					return editorDoneMsg{err}
				})
			}
		case "j", "down":
			m.modalScroll = clamp(m.modalScroll+1, 0, maxOffset)
		case "k", "up":
			m.modalScroll = clamp(m.modalScroll-1, 0, maxOffset)
		case "pgdown", "d":
			m.modalScroll = clamp(m.modalScroll+contentH/2, 0, maxOffset)
		case "pgup", "u":
			m.modalScroll = clamp(m.modalScroll-contentH/2, 0, maxOffset)
		case "home", "g":
			m.modalScroll = 0
		case "end", "G":
			m.modalScroll = maxOffset
		}
		return m, nil
	}

	// ── Main view ─────────────────────────────────────────────────────────────

	// New skill input
	if m.newSkillActive {
		switch key {
		case "esc":
			m.newSkillActive = false
			m.newSkillName = ""
		case "enter":
			name := strings.TrimSpace(m.newSkillName)
			m.newSkillActive = false
			m.newSkillName = ""
			if name != "" {
				skillPath, err := actions.CreateSkill(name)
				if err != nil {
					return m.setNotify("Failed: " + err.Error())
				}
				cmd := actions.EditorCmd(skillPath, 0)
				return m, tea.Batch(
					tea.ExecProcess(cmd, func(err error) tea.Msg { return editorDoneMsg{err} }),
				)
			}
		case "backspace":
			if len(m.newSkillName) > 0 {
				m.newSkillName = m.newSkillName[:len(m.newSkillName)-1]
			}
		default:
			if len(msg.Runes) == 1 {
				m.newSkillName += string(msg.Runes)
			}
		}
		return m, nil
	}

	// Confirm delete
	if m.confirmItem != nil {
		switch key {
		case "y":
			item := m.confirmItem
			m.confirmItem = nil
			var notifyMsg string
			switch v := item.(type) {
			case data.OcSkill:
				if err := actions.DeleteSkill(v); err != nil {
					return m.setNotify("Failed: " + err.Error())
				}
				notifyMsg = fmt.Sprintf(`Deleted skill "%s"`, v.ItemName())
			case data.OcCronJob:
				if _, err := actions.DeleteCronJob(v); err != nil {
					return m.setNotify("Failed: " + err.Error())
				}
				notifyMsg = fmt.Sprintf(`Deleted cron job "%s"`, v.ItemName())
			}
			m2, _ := m.setNotify(notifyMsg)
			m2m := m2.(Model)
			return m2m, loadDataCmd()
		default:
			m.confirmItem = nil
		}
		return m, nil
	}

	// Search active — handle before the h/esc panel-back so Esc cancels search first
	if m.searchActive {
		switch key {
		case "esc":
			m.searchActive = false
			m.searchQuery = ""
			m.itemIdx = 0
		case "enter":
			m.searchActive = false
		case "backspace":
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.itemIdx = 0
			}
		default:
			if len(msg.Runes) == 1 {
				m.searchQuery += string(msg.Runes)
				m.itemIdx = 0
			}
		}
		return m, nil
	}

	// Global keys (work in both panels)
	switch key {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "r":
		m.reloading = true
		return m, loadDataCmd()
	}

	// ── Categories panel ──────────────────────────────────────────────────────
	if m.panel == panelCategories {
		switch key {
		case "j", "down":
			m.catIdx = clamp(m.catIdx+1, 0, len(categoryOrder)-1)
			m.itemIdx = 0
			m.searchQuery = ""
		case "k", "up":
			m.catIdx = clamp(m.catIdx-1, 0, len(categoryOrder)-1)
			m.itemIdx = 0
			m.searchQuery = ""
		case "enter", "right", "tab":
			// go deeper into items
			m.panel = panelItems
		}
		return m, nil
	}

	// ── Items panel ───────────────────────────────────────────────────────────
	switch key {
	case "left", "esc":
		// go back to categories
		m.panel = panelCategories
		m.searchActive = false
		m.searchQuery = ""
		return m, nil
	case "/":
		m.searchActive = true
		return m, nil
	case "s":
		if categoryOrder[m.catIdx] == "skills" {
			switch m.scope {
			case "all":
				m.scope = "built-in"
			case "built-in":
				m.scope = "installed"
			default:
				m.scope = "all"
			}
			m.itemIdx = 0
		}
		return m, nil
	}

	items := m.filteredItems()
	switch key {
	case "j", "down":
		m.itemIdx = clamp(m.itemIdx+1, 0, len(items)-1)
	case "k", "up":
		m.itemIdx = clamp(m.itemIdx-1, 0, len(items)-1)
	case "n":
		if categoryOrder[m.catIdx] == "skills" {
			m.newSkillActive = true
			m.newSkillName = ""
			m.panel = panelItems
		}
	case "d":
		if m.itemIdx < len(items) {
			item := items[m.itemIdx]
			switch v := item.(type) {
			case data.OcSkill:
				if v.Scope == "installed" {
					m.confirmItem = item
				} else {
					return m.setNotify("Cannot delete built-in skills")
				}
			case data.OcCronJob:
				m.confirmItem = item
			}
		}
	case "o":
		if m.itemIdx < len(items) {
			item := items[m.itemIdx]
			path := actions.GetEditableFilePath(item)
			if path != "" {
				cmd := actions.EditorCmd(path, actions.GetEditLineNumber(item))
				return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
					return editorDoneMsg{err}
				})
			} else {
				return m.setNotify("No file to open for this item")
			}
		}
	case "t":
		if m.itemIdx < len(items) {
			item := items[m.itemIdx]
			var ok bool
			var err error
			switch v := item.(type) {
			case data.OcHook:
				ok, err = actions.ToggleHook(v)
			case data.OcCronJob:
				ok, err = actions.ToggleCron(v)
			case data.OcWebhook:
				ok, err = actions.ToggleWebhook(v)
			default:
				return m.setNotify("Toggle only works on hooks, cron jobs, and webhooks")
			}
			if err != nil {
				return m.setNotify("Failed: " + err.Error())
			}
			if ok {
				return m, loadDataCmd()
			}
		}
	case "enter", "right":
		if m.itemIdx < len(items) {
			item := items[m.itemIdx]
			if sess, ok := item.(data.OcSession); ok {
				s := sess
				m.transcriptSession = &s
				m.transcriptTurns = data.ParseTranscript(s.SessionFile)
				m.transcriptCursor = 0
				m.transcriptOffset = 0
				m.transcriptExpanded = map[int]bool{}
				m.mode = viewTranscript
			} else {
				m.modalItem = item
				m.modalScroll = 0
				m.fetchState = "idle"
				m.fetchedChanges = nil
				m.fetchedFixes = nil
				m.mode = viewDetail
				// Trigger changelog fetch for update items with no local content
				if upd, ok := item.(data.OcUpdateRelease); ok {
					if len(upd.Changes) == 0 && len(upd.Fixes) == 0 {
						m.fetchState = "fetching"
						return m, fetchChangelogCmd(upd.Version)
					}
				}
			}
		}
	}
	return m, nil
}

// ─── View ─────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	if m.width == 0 {
		return "loading…"
	}
	switch m.mode {
	case viewDetail:
		return renderDetailModal(&m)
	case viewTranscript:
		return renderTranscriptView(&m)
	default:
		return m.renderMain()
	}
}

func (m Model) renderMain() string {
	panelHeight := clamp(m.height-3, 10, m.height-3)
	visibleCount := clamp(panelHeight-5, 4, panelHeight-5)

	titleLine := m.renderTitle()
	statusLine := m.renderStatusLine()
	catPanel := m.renderCategoryPanel(panelHeight)
	itemPanel := m.renderItemPanel(panelHeight, visibleCount)
	panels := lipgloss.JoinHorizontal(lipgloss.Top, catPanel, itemPanel)
	statusBar := m.renderStatusBar()

	return titleLine + "\n" + statusLine + "\n" + panels + "\n" + statusBar
}

func (m Model) renderTitle() string {
	title := styleCyan.Bold(true).Render("ocview ") +
		styleGray.Render("OpenClaw Browser")
	if m.scope != "all" && categoryOrder[m.catIdx] == "skills" {
		title += styleGray.Render("  · ") + styleYellow.Render("scope: "+m.scope)
	}
	return " " + title
}

func (m Model) renderStatusLine() string {
	if m.notification != "" {
		return " " + styleGreen.Render(m.notification)
	}
	if m.reloading {
		return " " + styleYellow.Render("↻ reloading…")
	}
	if m.status == nil {
		return " " + styleDim.Render("checking service…")
	}
	s := m.status
	activeColor := colorYellow
	switch s.Active {
	case "running":
		activeColor = colorGreen
	case "failed":
		activeColor = colorRed
	}
	parts := []string{
		lipgloss.NewStyle().Foreground(activeColor).Render("● " + s.Active),
	}
	if s.Since != "" {
		parts = append(parts, styleGray.Render("since "+s.Since))
	}
	if s.SocketHealth != "unknown" {
		sc := colorGray
		if s.SocketHealth == "ok" {
			sc = colorGreen
		} else if s.SocketHealth == "stale" {
			sc = colorYellow
		}
		parts = append(parts, lipgloss.NewStyle().Foreground(sc).Render("socket: "+s.SocketHealth))
	}
	if s.Version != "" {
		parts = append(parts, styleGray.Faint(true).Render("v"+s.Version))
	}
	// Update available badge
	for _, upd := range m.appData.Updates {
		if upd.IsAvailable {
			parts = append(parts, styleYellow.Render("⬆ "+upd.Version+" available"))
			break
		}
	}
	return " " + strings.Join(parts, "  ")
}

func (m Model) renderCategoryPanel(height int) string {
	active := m.panel == panelCategories
	borderColor := colorGray
	if active {
		borderColor = colorWhite
	}

	headerColor := colorGray
	if active {
		headerColor = colorWhite
	}

	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Foreground(headerColor).Bold(true).Render(" OPENCLAW") + "\n")
	sb.WriteString("\n")

	for i, kind := range categoryOrder {
		items := m.itemsForCategory(kind)
		count := len(items)
		label := categoryLabel[kind]
		isSelected := i == m.catIdx

		line := fmt.Sprintf("%s%-12s%3d", map[bool]string{true: "▶ ", false: "  "}[isSelected], label, count)
		var rendered string
		if isSelected {
			rendered = styleSelected.Render(line)
		} else if active {
			rendered = styleWhite.Render(line)
		} else {
			rendered = styleGray.Render(line)
		}
		sb.WriteString(rendered + "\n")
	}

	// Pad to fill height (border=2, padding=2, header=2, categories=11)
	used := 2 + len(categoryOrder)
	for i := used; i < height-2; i++ {
		sb.WriteString("\n")
	}

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		Width(20).
		Height(height - 2).
		PaddingLeft(1).PaddingRight(1).
		Render(sb.String())
}

func (m Model) renderItemPanel(height, visibleCount int) string {
	active := m.panel == panelItems
	borderColor := colorGray
	if active {
		borderColor = colorWhite
	}
	headerColor := colorGray
	if active {
		headerColor = colorWhite
	}

	kind := categoryOrder[m.catIdx]
	items := m.filteredItems()
	itemPanelWidth := m.width - 24 // 22 cat + 2 for borders overlap
	if itemPanelWidth < 20 {
		itemPanelWidth = 20
	}
	innerWidth := itemPanelWidth - 4 // border(2) + padding(2)

	var sb strings.Builder

	// Header row
	headerLabel := fmt.Sprintf(" %s (%d)", categoryPanelLabel[kind], len(items))
	searchHint := ""
	if !m.searchActive {
		searchHint = styleGray.Render(" /search ")
	}
	headerLeft := lipgloss.NewStyle().Foreground(headerColor).Bold(true).Render(headerLabel)
	headerLine := lipgloss.NewStyle().Width(innerWidth).Render(
		lipgloss.JoinHorizontal(lipgloss.Bottom,
			lipgloss.NewStyle().Width(innerWidth-len(searchHint)).Render(headerLeft),
			searchHint,
		),
	)
	sb.WriteString(headerLine + "\n")

	// Input row
	if m.newSkillActive {
		sb.WriteString(styleGreen.Render("+ ") + m.newSkillName + "█\n")
	} else if m.searchActive {
		sb.WriteString(styleYellow.Render("/") + " " + m.searchQuery + "█\n")
	} else {
		sb.WriteString("\n")
	}

	if len(items) == 0 {
		sb.WriteString(styleGray.Render("  (none)") + "\n")
	}

	// Scroll offset
	scrollOffset := clamp(m.itemIdx-visibleCount/2, 0, clamp(len(items)-visibleCount, 0, len(items)))
	visible := items
	if scrollOffset < len(items) {
		visible = items[scrollOffset:]
	}
	if len(visible) > visibleCount {
		visible = visible[:visibleCount]
	}

	nameWidth := 31

	for vi, item := range visible {
		realIdx := vi + scrollOffset
		isSelected := realIdx == m.itemIdx

		name := truncate(item.ItemName(), 30)
		arrow := "  "
		if isSelected {
			arrow = "▶ "
		}
		namePart := arrow + padRight(name, nameWidth)

		var extra string
		switch v := item.(type) {
		case data.OcSkill:
			scopeColor := colorCyan
			scopeText := "[built-in]"
			if v.Scope == "installed" {
				scopeColor = colorGreen
				scopeText = "[installed]"
			}
			extra = " " + lipgloss.NewStyle().Foreground(scopeColor).Render(scopeText)
		case data.OcMemoryChunk:
			preview := strings.ReplaceAll(v.Text, "\n", " ")
			if len([]rune(preview)) > 40 {
				preview = string([]rune(preview)[:40])
			}
			extra = " " + styleGray.Render(preview)
		case data.OcUpdateRelease:
			badge := ""
			badgeColor := colorGray
			if v.IsInstalled {
				badge = "[installed]"
				badgeColor = colorGreen
			} else if v.IsAvailable {
				badge = "[available]"
				badgeColor = colorYellow
			}
			counts := ""
			if v.ChangeCount > 0 {
				counts = fmt.Sprintf(" %dc %df", len(v.Changes), len(v.Fixes))
			}
			if badge != "" {
				extra += " " + lipgloss.NewStyle().Foreground(badgeColor).Render(badge)
			}
			if counts != "" {
				extra += styleGray.Render(counts)
			}
		case data.OcAuditEntry:
			t, _ := time.Parse(time.RFC3339, v.TS)
			ts := t.Format("Jan 2 15:04")
			hasSusp := len(v.Suspicious) > 0
			if hasSusp {
				extra = "  " + styleRed.Render(ts+"  ⚠")
			} else {
				extra = "  " + styleGray.Render(ts)
			}
		case data.OcTaskRun:
			statusColor := colorGray
			switch v.Status {
			case "succeeded":
				statusColor = colorGreen
			case "failed", "error":
				statusColor = colorRed
			case "running":
				statusColor = colorYellow
			}
			badge := lipgloss.NewStyle().Foreground(statusColor).Render("[" + v.Status + "]")
			extra = "  " + badge
			if v.Runtime != "" && v.Runtime != "cron" {
				extra += " " + styleGray.Render(v.Runtime)
			}
		case data.OcDevice:
			parts := ""
			if v.Platform != "" {
				parts = v.Platform
			}
			if v.Role != "" {
				if parts != "" {
					parts += "  "
				}
				parts += v.Role
			}
			statusColor := colorGreen
			if v.Status == "pending" {
				statusColor = colorYellow
			}
			extra = "  " + lipgloss.NewStyle().Foreground(statusColor).Render("["+v.Status+"]")
			if parts != "" {
				extra += " " + styleGray.Render(parts)
			}
		case data.OcConfigSection:
			if v.Summary != "" {
				extra = "  " + styleGray.Render(truncate(v.Summary, 40))
			}
		case data.OcLogFile:
			extra = "  " + styleGray.Render(fmt.Sprintf("%d lines", v.LineCount))
		}

		if isSelected {
			line := styleSelected.Render(namePart) + extra
			sb.WriteString(line + "\n")
		} else if active {
			sb.WriteString(styleWhite.Render(namePart) + extra + "\n")
		} else {
			sb.WriteString(styleGray.Render(namePart) + extra + "\n")
		}
	}

	// Scroll indicator
	if len(items) > visibleCount {
		indicator := fmt.Sprintf(" ↑↓ %d–%d/%d",
			scrollOffset+1,
			clamp(scrollOffset+visibleCount, 0, len(items)),
			len(items))
		sb.WriteString(styleGray.Render(indicator) + "\n")
	}

	// Pad remaining lines
	lines := strings.Count(sb.String(), "\n")
	for lines < height-2 {
		sb.WriteString("\n")
		lines++
	}

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		Width(itemPanelWidth - 4).
		Height(height - 2).
		PaddingLeft(1).PaddingRight(1).
		Render(sb.String())
}

func (m Model) renderStatusBar() string {
	if m.confirmItem != nil {
		label := m.confirmItem.ItemName()
		return " " + styleRed.Render(fmt.Sprintf(`Delete "%s"?`, label)) +
			"   " + styleYellow.Bold(true).Render("y") + styleGray.Render(" confirm") +
			"   " + styleYellow.Bold(true).Render("n/Esc") + styleGray.Render(" cancel")
	}
	if m.newSkillActive {
		return " " + styleGray.Render("Skill directory name") +
			"   " + styleYellow.Bold(true).Render("Esc") + styleGray.Render(" cancel") +
			"   " + styleYellow.Bold(true).Render("↵") + styleGray.Render(" create")
	}
	if m.searchActive {
		return " " + styleGray.Render("Type to filter") +
			"   " + styleYellow.Bold(true).Render("Esc") + styleGray.Render(" cancel") +
			"   " + styleYellow.Bold(true).Render("↵") + styleGray.Render(" confirm")
	}

	items := m.filteredItems()
	var selectedItem data.AnyItem
	if m.itemIdx < len(items) {
		selectedItem = items[m.itemIdx]
	}

	canEdit := false
	canToggle := false
	canDelete := false
	isSession := categoryOrder[m.catIdx] == "sessions"

	if selectedItem != nil {
		switch v := selectedItem.(type) {
		case data.OcSkill:
			canEdit = true
			canDelete = v.Scope == "installed"
		case data.OcWorkspaceFile:
			canEdit = true
		case data.OcMemoryChunk:
			canEdit = true
		case data.OcHook:
			_ = v
			canEdit = true
			canToggle = true
		case data.OcCronJob:
			canEdit = true
			canToggle = true
			canDelete = true
		case data.OcWebhook:
			canEdit = true
			canToggle = true
		case data.OcMcpServer:
			_ = v
			canEdit = true
		case data.OcModel:
			_ = v
			canEdit = true
		}
	}

	k := func(key, label string) string {
		return styleYellow.Bold(true).Render(key) + styleGray.Render(":"+label)
	}

	var parts []string
	if m.panel == panelCategories {
		parts = []string{
			k("↑↓", "nav"),
			k("↵/→/Tab", "open"),
			k("r", "reload"),
			k("q", "quit"),
		}
	} else {
		parts = []string{
			k("↑↓", "nav"),
			k("←/Esc", "back"),
			k("/", "search"),
		}
		if categoryOrder[m.catIdx] == "skills" {
			parts = append(parts, k("s", "scope"), k("n", "new"))
		}
		if isSession {
			parts = append(parts, k("↵/→", "transcript"))
		} else {
			parts = append(parts, k("↵/→", "detail"))
		}
		if canEdit {
			parts = append(parts, k("o", "edit"))
		}
		if canToggle {
			parts = append(parts, k("t", "toggle"))
		}
		if canDelete {
			parts = append(parts, k("d", "delete"))
		}
		parts = append(parts, k("r", "reload"), k("q", "quit"))
	}

	return " " + strings.Join(parts, "  ")
}

// ─── Data helpers ─────────────────────────────────────────────────────────────

func (m Model) itemsForCategory(kind string) []data.AnyItem {
	var items []data.AnyItem
	switch kind {
	case "skills":
		for _, s := range m.appData.Skills {
			items = append(items, s)
		}
	case "hooks":
		for _, h := range m.appData.Hooks {
			items = append(items, h)
		}
	case "models":
		for _, mo := range m.appData.Models {
			items = append(items, mo)
		}
	case "workspace":
		for _, w := range m.appData.Workspace {
			items = append(items, w)
		}
	case "mcp":
		for _, mc := range m.appData.MCP {
			items = append(items, mc)
		}
	case "sessions":
		for _, s := range m.appData.Sessions {
			items = append(items, s)
		}
	case "cron":
		for _, c := range m.appData.Cron {
			items = append(items, c)
		}
	case "memory":
		for _, mc := range m.appData.Memory {
			items = append(items, mc)
		}
	case "updates":
		for _, u := range m.appData.Updates {
			items = append(items, u)
		}
	case "webhooks":
		for _, w := range m.appData.Webhooks {
			items = append(items, w)
		}
	case "auditlog":
		for _, a := range m.appData.AuditLog {
			items = append(items, a)
		}
	case "tasks":
		for _, t := range m.appData.Tasks {
			items = append(items, t)
		}
	case "agentconfig":
		for _, c := range m.appData.AgentConfig {
			items = append(items, c)
		}
	case "devices":
		for _, d := range m.appData.Devices {
			items = append(items, d)
		}
	case "logs":
		for _, l := range m.appData.Logs {
			items = append(items, l)
		}
	}
	return items
}

func (m Model) filteredItems() []data.AnyItem {
	kind := categoryOrder[m.catIdx]
	items := m.itemsForCategory(kind)

	// Scope filter (skills only)
	if kind == "skills" && m.scope != "all" {
		var filtered []data.AnyItem
		for _, item := range items {
			if s, ok := item.(data.OcSkill); ok && s.Scope == m.scope {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	// Search filter
	if m.searchQuery != "" {
		q := strings.ToLower(m.searchQuery)
		var filtered []data.AnyItem
		for _, item := range items {
			name := strings.ToLower(item.ItemName())
			if strings.Contains(name, q) {
				filtered = append(filtered, item)
				continue
			}
			// Also search description and text fields
			switch v := item.(type) {
			case data.OcSkill:
				if strings.Contains(strings.ToLower(v.Description), q) {
					filtered = append(filtered, item)
				}
			case data.OcHook:
				if strings.Contains(strings.ToLower(v.Description), q) {
					filtered = append(filtered, item)
				}
			case data.OcMemoryChunk:
				if strings.Contains(strings.ToLower(v.Text), q) {
					filtered = append(filtered, item)
				}
			}
		}
		items = filtered
	}
	return items
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (m Model) setNotify(msg string) (tea.Model, tea.Cmd) {
	m.notification = msg
	m.notifySeq++
	return m, notifyClearCmd(m.notifySeq)
}

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
