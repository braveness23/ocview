package ui

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/braveness23/ocview/internal/data"
	"github.com/charmbracelet/lipgloss"
)

// renderDetailLines returns the full scrollable content for a detail modal as
// a flat []string, one string per display line.
func renderDetailLines(item data.AnyItem, textWidth int, fetchedChanges, fetchedFixes []string, fetchState string) []string {
	switch v := item.(type) {
	case data.OcSkill:
		return wrapText(v.FullContent, textWidth)
	case data.OcWorkspaceFile:
		return workspaceLines(v, textWidth)
	case data.OcMemoryChunk:
		return memoryLines(v, textWidth)
	case data.OcUpdateRelease:
		return updateLines(v, textWidth, fetchedChanges, fetchedFixes, fetchState)
	case data.OcAuditEntry:
		return auditLines(v, textWidth)
	}
	return nil
}

func workspaceLines(v data.OcWorkspaceFile, w int) []string {
	lines := []string{
		fmt.Sprintf("path:     %s", shortPath(v.FilePath)),
		fmt.Sprintf("words:    %d", v.WordCount),
		fmt.Sprintf("modified: %s", fmtTimeISO(v.LastModified)),
		"",
		divider(w),
		"",
	}
	return append(lines, wrapText(v.FullContent, w)...)
}

func memoryLines(v data.OcMemoryChunk, w int) []string {
	lines := []string{
		fmt.Sprintf("source:   %s", v.Source),
		fmt.Sprintf("path:     %s", shortPath(v.Path)),
		fmt.Sprintf("lines:    %d–%d", v.StartLine, v.EndLine),
		fmt.Sprintf("model:    %s", v.Model),
		fmt.Sprintf("updated:  %s", fmtTimestamp(v.UpdatedAt)),
		"",
		divider(w),
		"",
	}
	return append(lines, wrapText(v.Text, w)...)
}

func updateLines(v data.OcUpdateRelease, w int, fetchedChanges, fetchedFixes []string, fetchState string) []string {
	changes := v.Changes
	fixes := v.Fixes
	if len(changes) == 0 {
		changes = fetchedChanges
	}
	if len(fixes) == 0 {
		fixes = fetchedFixes
	}

	var result []string
	result = append(result, fmt.Sprintf("version:  %s", v.Version))

	var statusParts []string
	if v.IsInstalled {
		statusParts = append(statusParts, "● installed")
	}
	if v.IsAvailable {
		statusParts = append(statusParts, "↑ available")
	}
	if v.IsLatest {
		statusParts = append(statusParts, "latest")
	}
	if len(statusParts) > 0 {
		result = append(result, fmt.Sprintf("status:   %s", strings.Join(statusParts, "  ")))
	}
	if v.InstallRecord != nil {
		ts := v.InstallRecord.Timestamp
		if t, err := time.Parse(time.RFC3339, ts); err == nil {
			ts = t.Format("Jan 2, 2006 03:04 PM")
		}
		result = append(result, fmt.Sprintf("installed: %s  (from %s)", ts, v.InstallRecord.From))
	}
	if v.LastChecked != "" && v.IsInstalled {
		ts := v.LastChecked
		if t, err := time.Parse(time.RFC3339, ts); err == nil {
			ts = t.Format("Jan 2, 3:04 PM")
		}
		result = append(result, fmt.Sprintf("checked:  %s", ts))
	}
	result = append(result, "", divider(w))

	switch fetchState {
	case "fetching":
		result = append(result, "", "Fetching changelog from GitHub…")
	case "error":
		result = append(result, "", "(could not fetch changelog from GitHub)")
	default:
		if len(changes) == 0 && len(fixes) == 0 {
			result = append(result, "", "(no changelog entry available)")
		} else {
			if len(changes) > 0 {
				result = append(result, "", "### Changes", "")
				for _, c := range changes {
					wrapped := wrapText("• "+c, w-2)
					for i, l := range wrapped {
						if i > 0 {
							l = "  " + l
						}
						result = append(result, l)
					}
				}
			}
			if len(fixes) > 0 {
				result = append(result, "", "### Fixes", "")
				for _, f := range fixes {
					wrapped := wrapText("• "+f, w-2)
					for i, l := range wrapped {
						if i > 0 {
							l = "  " + l
						}
						result = append(result, l)
					}
				}
			}
		}
	}
	return result
}

func auditLines(v data.OcAuditEntry, w int) []string {
	ts := v.TS
	if t, err := time.Parse(time.RFC3339, v.TS); err == nil {
		ts = t.Format("Jan 2, 2006  15:04:05")
	}

	prevBytes := "(new)"
	if v.PreviousBytes != nil {
		prevBytes = fmt.Sprintf("%d", *v.PreviousBytes)
	}

	result := []string{
		fmt.Sprintf("time:      %s", ts),
		fmt.Sprintf("event:     %s", v.Event),
		fmt.Sprintf("source:    %s", v.Source),
		fmt.Sprintf("result:    %s", v.Result),
		fmt.Sprintf("pid:       %d", v.PID),
		"",
		divider(w),
		"",
		"### Config file",
		"",
		fmt.Sprintf("path:      %s", v.ConfigPath),
		fmt.Sprintf("size:      %s → %d bytes", prevBytes, v.NextBytes),
		fmt.Sprintf("hash:      %s… → %s…",
			truncate(v.PreviousHash, 16), truncate(v.NextHash, 16)),
	}
	if v.GatewayModeBefore != nil || v.GatewayModeAfter != nil {
		before := "(none)"
		after := "(none)"
		if v.GatewayModeBefore != nil {
			before = *v.GatewayModeBefore
		}
		if v.GatewayModeAfter != nil {
			after = *v.GatewayModeAfter
		}
		result = append(result, fmt.Sprintf("mode:      %s → %s", before, after))
	}
	result = append(result, "", divider(w), "", "### Command", "")
	result = append(result, wrapText(strings.Join(v.Argv, " "), w)...)
	if len(v.Suspicious) > 0 {
		result = append(result, "", divider(w), "", "### ⚠ Suspicious activity", "")
		for _, s := range v.Suspicious {
			result = append(result, "• "+s)
		}
	}
	return result
}

// isScrollableKind returns true for item kinds that use the scrollable text renderer.
func isScrollableKind(kind string) bool {
	switch kind {
	case "skill", "workspace", "memory", "update", "auditlog":
		return true
	}
	return false
}

// renderDetailBody returns the rendered (non-scrollable) body for a detail modal.
// Used only for non-scrollable kinds.
func renderDetailBody(item data.AnyItem, textWidth int) string {
	var sb strings.Builder
	switch v := item.(type) {
	case data.OcHook:
		renderHookBody(&sb, v, textWidth)
	case data.OcModel:
		renderModelBody(&sb, v)
	case data.OcMcpServer:
		renderMcpBody(&sb, v)
	case data.OcSession:
		renderSessionBody(&sb, v)
	case data.OcCronJob:
		renderCronBody(&sb, v)
	case data.OcWebhook:
		renderWebhookBody(&sb, v)
	}
	return sb.String()
}

func field(sb *strings.Builder, label, value string) {
	sb.WriteString(fmt.Sprintf("  %-18s%s\n", label, value))
}

func section(sb *strings.Builder, title string) {
	sb.WriteString("\n")
	sb.WriteString(styleGray.Render("  "+title) + "\n")
}

func renderHookBody(sb *strings.Builder, v data.OcHook, textWidth int) {
	if v.Description != "" {
		sb.WriteString("\n  " + v.Description + "\n")
	}
	section(sb, "STATUS")
	enabled := styleRed.Render("no")
	if v.Enabled {
		enabled = styleGreen.Render("yes")
	}
	sb.WriteString(fmt.Sprintf("  %-18s%s\n", "enabled", enabled))

	section(sb, "CONFIG  (from openclaw.json → hooks.internal.entries)")
	b, _ := json.MarshalIndent(v.RawConfig, "  ", "  ")
	sb.WriteString("\n  " + styleGray.Render(string(b)) + "\n")

	section(sb, "IMPLEMENTATION")
	npmRoot := "(run: npm root -g)"
	if out, err := exec.Command("npm", "root", "-g").Output(); err == nil {
		npmRoot = strings.TrimSpace(string(out)) + "/openclaw"
	}
	sb.WriteString("\n  " + styleYellow.Render("Built-in hook — logic lives inside the openclaw npm package.") + "\n")
	sb.WriteString("  Source: " + styleCyan.Render(npmRoot) + "\n")
}

func renderModelBody(sb *strings.Builder, v data.OcModel) {
	section(sb, "IDENTITY")
	field(sb, "provider", styleCyan.Render(v.Provider))
	field(sb, "model id", v.ID)
	reasoning := styleGray.Render("no")
	if v.Reasoning {
		reasoning = styleGreen.Render("yes")
	}
	field(sb, "reasoning", reasoning)

	section(sb, "LIMITS")
	field(sb, "context", fmtNumber(v.ContextWindow)+" tokens")
	field(sb, "max output", fmtNumber(v.MaxTokens)+" tokens")

	section(sb, "COST  (per million tokens)")
	field(sb, "input", fmtCost(v.CostInput))
	field(sb, "output", fmtCost(v.CostOutput))
}

func renderMcpBody(sb *strings.Builder, v data.OcMcpServer) {
	section(sb, "DETAILS")
	field(sb, "transport", styleCyan.Render(v.Transport))
	enabledStr := styleGray.Render("no")
	if v.Enabled {
		enabledStr = styleGreen.Render("yes")
	}
	field(sb, "enabled", enabledStr)
	availStr := styleRed.Render("no")
	if !v.Enabled {
		availStr = styleGray.Render("disabled")
	} else if v.Available {
		availStr = styleGreen.Render("yes")
	}
	field(sb, "available", availStr)
	if v.URL != "" {
		field(sb, "url", v.URL)
	}
	if v.Command != "" {
		parts := append([]string{v.Command}, v.Args...)
		field(sb, "command", styleCyan.Render(strings.Join(parts, " ")))
	}
	if len(v.Dependencies) > 0 {
		section(sb, "DEPENDENCIES")
		for _, dep := range v.Dependencies {
			mark := styleRed.Render("✗")
			if dep.Met {
				mark = styleGreen.Render("✓")
			}
			field(sb, mark, dep.Name)
		}
	}
	if len(v.Headers) > 0 {
		section(sb, "HEADERS")
		for k, val := range v.Headers {
			display := val
			if len(display) > 8 {
				display = display[:8] + "••••••••"
			}
			field(sb, k, display)
		}
	}
}

func renderSessionBody(sb *strings.Builder, v data.OcSession) {
	section(sb, "DETAILS")
	field(sb, "channel", styleCyan.Render(v.Channel))
	field(sb, "updated", fmtTimestamp(v.UpdatedAt))
	field(sb, "size", fmt.Sprintf("%d KB", v.SizeKb))
	field(sb, "file", shortPath(v.SessionFile))
	sb.WriteString("\n  " + styleGray.Render("(Press Enter on session in list to view transcript)") + "\n")
}

func renderCronBody(sb *strings.Builder, v data.OcCronJob) {
	if v.Description != "" {
		sb.WriteString("\n  " + v.Description + "\n")
	}
	section(sb, "STATUS")
	enabledStr := styleYellow.Render("no")
	if v.Enabled {
		enabledStr = styleGreen.Render("yes")
	}
	field(sb, "enabled", enabledStr)
	field(sb, "schedule", styleCyan.Render(v.Schedule))

	section(sb, "COMMAND")
	sb.WriteString("\n  " + v.Command + "\n")
}

func renderWebhookBody(sb *strings.Builder, v data.OcWebhook) {
	if v.Description != "" {
		sb.WriteString("\n  " + v.Description + "\n")
	}
	section(sb, "STATUS")
	enabledStr := styleYellow.Render("no")
	if v.Enabled {
		enabledStr = styleGreen.Render("yes")
	}
	field(sb, "enabled", enabledStr)

	section(sb, "ROUTING")
	field(sb, "path", styleCyan.Render(v.Path))
	field(sb, "sessionKey", v.SessionKey)
	field(sb, "controllerId", v.ControllerID)

	section(sb, "AUTH")
	field(sb, "secret", styleGray.Render(v.Secret))
}

// renderScrollableLines renders a visible slice of scrollable lines with markdown colouring.
func renderScrollableLines(lines []string, offset, height, textWidth int) string {
	visible := lines
	if offset < len(lines) {
		visible = lines[offset:]
	} else {
		visible = nil
	}
	if len(visible) > height {
		visible = visible[:height]
	}

	var sb strings.Builder
	for _, line := range visible {
		col, bold := mdLineStyle(line)
		st := lipgloss.NewStyle().Foreground(lipgloss.Color(col))
		if bold == "bold" {
			st = st.Bold(true)
		}
		// Truncate to textWidth
		displayed := line
		if len([]rune(displayed)) > textWidth {
			displayed = string([]rune(displayed)[:textWidth])
		}
		if displayed == "" {
			displayed = " "
		}
		sb.WriteString("  " + st.Render(displayed) + "\n")
	}
	// Pad remaining lines
	for i := len(visible); i < height; i++ {
		sb.WriteString("\n")
	}
	return sb.String()
}

// renderDetailModal renders the full modal as a string.
func renderDetailModal(m *Model) string {
	item := m.modalItem
	modalWidth := min(m.width, 90)
	textWidth := modalWidth - 6 // border(2) + padding(2) each side → 6
	chromeRows := 10
	contentHeight := clamp(m.height-chromeRows, 5, m.height-4)

	kLabel := kindLabel(item.ItemKind())
	kColor := kindColor(item.ItemKind())

	labelBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(kColor).
		Foreground(kColor).
		Bold(true).
		PaddingLeft(1).PaddingRight(1).
		Render(kLabel)

	titleLine := lipgloss.JoinHorizontal(lipgloss.Center,
		labelBox, "  ",
		lipgloss.NewStyle().Bold(true).Foreground(colorWhite).Render(truncate(item.ItemName(), textWidth-10)),
	)

	div := styleGray.Render("  " + divider(modalWidth-4))

	var bodyStr string
	scrollable := isScrollableKind(item.ItemKind())
	if scrollable {
		lines := renderDetailLines(item, textWidth, m.fetchedChanges, m.fetchedFixes, m.fetchState)
		totalLines := len(lines)
		maxOffset := clamp(totalLines-contentHeight, 0, totalLines)
		if m.modalScroll > maxOffset {
			m.modalScroll = maxOffset
		}
		bodyStr = renderScrollableLines(lines, m.modalScroll, contentHeight, textWidth)

		// Footer counter
		end := clamp(m.modalScroll+contentHeight, 0, totalLines)
		footerRight := ""
		if totalLines > contentHeight {
			footerRight = styleGray.Render(fmt.Sprintf("%d–%d/%d", m.modalScroll+1, end, totalLines))
		}
		footerLeft := buildDetailFooter(item, scrollable)
		footerLine := lipgloss.NewStyle().Width(modalWidth - 6).Render(
			lipgloss.JoinHorizontal(lipgloss.Bottom,
				lipgloss.NewStyle().Width(modalWidth-6-len(footerRight)-2).Render(footerLeft),
				footerRight,
			),
		)
		inner := titleLine + "\n" + div + "\n" + bodyStr + div + "\n" + footerLine
		return styleModal.Width(modalWidth - 2).Render(inner)
	}

	bodyStr = renderDetailBody(item, textWidth)
	// Pad body to contentHeight
	lines := strings.Split(bodyStr, "\n")
	for len(lines) < contentHeight {
		lines = append(lines, "")
	}
	bodyStr = strings.Join(lines[:contentHeight], "\n")

	footerLeft := buildDetailFooter(item, scrollable)
	inner := titleLine + "\n" + div + "\n" + bodyStr + "\n" + div + "\n" + footerLeft
	return styleModal.Width(modalWidth - 2).Render(inner)
}

func buildDetailFooter(item data.AnyItem, scrollable bool) string {
	y := func(k, label string) string {
		return styleYellow.Bold(true).Render(k) + styleGray.Render(" "+label)
	}
	keys := []string{y("q/Esc", "close")}
	kind := item.ItemKind()
	if kind == "skill" || kind == "workspace" || kind == "memory" {
		keys = append(keys, y("o", "edit"))
	}
	if scrollable {
		keys = append(keys, y("↑↓", "scroll"), y("PgDn/PgUp", "page"), y("Home/End", "top/btm"))
	}
	return strings.Join(keys, "   ")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getChangelogURL attempts to find the GitHub raw CHANGELOG.md URL from the installed package.
func getChangelogURL() string {
	out, err := exec.Command("npm", "root", "-g").Output()
	if err == nil {
		pkgPath := filepath.Join(strings.TrimSpace(string(out)), "openclaw", "package.json")
		if b, err2 := readFileSafe(pkgPath); err2 == nil {
			var pkg struct {
				Repository struct {
					URL string `json:"url"`
				} `json:"repository"`
			}
			if json.Unmarshal(b, &pkg) == nil {
				repoURL := strings.TrimPrefix(pkg.Repository.URL, "git+")
				repoURL = strings.TrimSuffix(repoURL, ".git")
				if strings.Contains(repoURL, "github.com") {
					parts := strings.Split(strings.TrimPrefix(repoURL, "https://github.com/"), "/")
					if len(parts) >= 2 {
						return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/CHANGELOG.md", parts[0], parts[1])
					}
				}
			}
		}
	}
	return "https://raw.githubusercontent.com/openclaw/openclaw/main/CHANGELOG.md"
}

func readFileSafe(path string) ([]byte, error) {
	return readFile(path)
}
