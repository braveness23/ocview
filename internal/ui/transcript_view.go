package ui

import (
	"fmt"
	"strings"

	"github.com/braveness23/ocview/internal/data"
	"github.com/charmbracelet/lipgloss"
)

type displayLine struct {
	text        string
	color       lipgloss.Color
	bold        bool
	dim         bool
	turnIdx     int
	isToolHeader bool
}

func buildTranscriptLines(turns []data.Turn, expanded map[int]bool, width int) []displayLine {
	var lines []displayLine

	for idx, turn := range turns {
		switch turn.Kind {
		case data.TurnText:
			roleLabel := turn.Role
			roleColor := colorWhite
			switch turn.Role {
			case "user":
				roleColor = colorCyan
			case "assistant":
				roleColor = colorWhite
			case "system":
				roleColor = colorMagenta
			}
			pad := clamp(width-len(roleLabel)-4, 0, width)
			lines = append(lines,
				displayLine{text: "", color: colorGray, dim: true, turnIdx: idx},
				displayLine{
					text:    fmt.Sprintf("── %s %s", roleLabel, divider(pad)),
					color:   colorGray, dim: true, turnIdx: idx,
				},
			)
			if strings.TrimSpace(turn.Text) == "" {
				continue
			}
			for _, tl := range wrapText(turn.Text, width-2) {
				lines = append(lines, displayLine{
					text: "  " + tl, color: roleColor, turnIdx: idx,
				})
			}

		case data.TurnToolCall:
			exp := expanded[idx]
			arrow := "▶"
			if exp {
				arrow = "▼"
			}
			lines = append(lines, displayLine{
				text:         fmt.Sprintf("  %s [tool] %s", arrow, turn.ToolName),
				color:        colorYellow,
				bold:         true,
				turnIdx:      idx,
				isToolHeader: true,
			})
			if exp {
				for _, il := range wrapText(turn.Input, width-4) {
					lines = append(lines, displayLine{
						text: "    " + il, color: colorGray, turnIdx: idx,
					})
				}
			}

		case data.TurnToolResult:
			exp := expanded[idx]
			arrow := "▶"
			if exp {
				arrow = "▼"
			}
			nameLabel := turn.ToolName
			if nameLabel == "" {
				nameLabel = truncate(turn.ToolUseID, 8)
			}
			lines = append(lines, displayLine{
				text:         fmt.Sprintf("  %s [result] %s", arrow, nameLabel),
				color:        colorGray,
				dim:          true,
				turnIdx:      idx,
				isToolHeader: true,
			})
			if exp {
				preview := turn.Content
				if len(preview) > 3000 {
					preview = preview[:3000] + "\n…(truncated)"
				}
				for _, rl := range wrapText(preview, width-4) {
					lines = append(lines, displayLine{
						text: "    " + rl, color: colorGray, dim: true, turnIdx: idx,
					})
				}
			}
		}
	}
	return lines
}

func renderTranscriptView(m *Model) string {
	sess := m.transcriptSession
	width := m.width
	contentHeight := clamp(m.height-4, 4, m.height-4) // 3 header + 1 footer
	textWidth := width - 2

	div := strings.Repeat("─", width)

	// Header
	header := lipgloss.JoinHorizontal(lipgloss.Center,
		styleCyan.Bold(true).Render("ocview"),
		"  ",
		lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(colorWhite).
			Foreground(colorWhite).Bold(true).PaddingLeft(1).PaddingRight(1).Render("session"),
		"  ",
		lipgloss.NewStyle().Bold(true).Foreground(colorWhite).Render(sess.Name_),
	)
	meta := styleGray.Render(fmt.Sprintf(" %s  ·  %d KB  ·  %s",
		fmtTimestamp(sess.UpdatedAt), sess.SizeKb, shortPath(sess.SessionFile)))

	// Build lines
	displayLines := buildTranscriptLines(m.transcriptTurns, m.transcriptExpanded, textWidth)
	totalLines := len(displayLines)

	// Render content
	var contentSB strings.Builder
	for i := 0; i < contentHeight; i++ {
		absIdx := m.transcriptOffset + i
		if absIdx >= totalLines {
			contentSB.WriteString("\n")
			continue
		}
		dl := displayLines[absIdx]
		isCursor := absIdx == m.transcriptCursor

		text := dl.text
		if text == "" {
			text = " "
		}
		// Truncate to width
		runes := []rune(text)
		if len(runes) > width {
			text = string(runes[:width])
		}

		var st lipgloss.Style
		if isCursor {
			st = lipgloss.NewStyle().Background(colorCyan).Foreground(colorBlack).Bold(true)
		} else {
			st = lipgloss.NewStyle().Foreground(dl.color)
			if dl.bold {
				st = st.Bold(true)
			}
			if dl.dim {
				st = st.Faint(true)
			}
		}
		contentSB.WriteString(lipgloss.NewStyle().Width(width).Render(st.Render(text)) + "\n")
	}

	// Footer
	hasTools := false
	for _, t := range m.transcriptTurns {
		if t.Kind == data.TurnToolCall || t.Kind == data.TurnToolResult {
			hasTools = true
			break
		}
	}
	footerKeys := []string{
		styleYellow.Bold(true).Render("j/k") + styleGray.Render(" scroll"),
	}
	if hasTools {
		footerKeys = append(footerKeys, styleYellow.Bold(true).Render("↵")+styleGray.Render(" expand"))
	}
	footerKeys = append(footerKeys,
		styleYellow.Bold(true).Render("d/u")+styleGray.Render(" page"),
		styleYellow.Bold(true).Render("g/G")+styleGray.Render(" top/btm"),
		styleYellow.Bold(true).Render("q")+styleGray.Render(" back"),
	)
	footerLeft := " " + strings.Join(footerKeys, "   ")
	footerRight := ""
	if totalLines > 0 {
		footerRight = styleGray.Render(fmt.Sprintf("line %d/%d ", m.transcriptCursor+1, totalLines))
	}
	footerLine := lipgloss.NewStyle().Width(width).Render(
		lipgloss.JoinHorizontal(lipgloss.Bottom,
			lipgloss.NewStyle().Width(width-len(footerRight)).Render(footerLeft),
			footerRight,
		),
	)

	return " " + header + "\n" +
		meta + "\n" +
		styleGray.Render(div) + "\n" +
		contentSB.String() +
		footerLine
}
