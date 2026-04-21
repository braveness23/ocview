package ui

import "github.com/charmbracelet/lipgloss"

var (
	colorCyan    = lipgloss.Color("14")
	colorGreen   = lipgloss.Color("10")
	colorYellow  = lipgloss.Color("11")
	colorRed     = lipgloss.Color("9")
	colorBlue    = lipgloss.Color("12")
	colorMagenta = lipgloss.Color("13")
	colorWhite   = lipgloss.Color("15")
	colorGray    = lipgloss.Color("8")
	colorBlack   = lipgloss.Color("0")

	styleBold  = lipgloss.NewStyle().Bold(true)
	styleDim   = lipgloss.NewStyle().Faint(true)
	styleGray  = lipgloss.NewStyle().Foreground(colorGray)
	styleWhite = lipgloss.NewStyle().Foreground(colorWhite)
	styleCyan  = lipgloss.NewStyle().Foreground(colorCyan)
	styleGreen = lipgloss.NewStyle().Foreground(colorGreen)
	styleYellow = lipgloss.NewStyle().Foreground(colorYellow)
	styleRed   = lipgloss.NewStyle().Foreground(colorRed)

	styleSelected = lipgloss.NewStyle().
			Background(colorCyan).
			Foreground(colorBlack).
			Bold(true)

	styleCatPanel = func(active bool) lipgloss.Style {
		borderColor := colorGray
		if active {
			borderColor = colorWhite
		}
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(borderColor).
			PaddingLeft(1).PaddingRight(1)
	}

	styleItemPanel = func(active bool) lipgloss.Style {
		borderColor := colorGray
		if active {
			borderColor = colorWhite
		}
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(borderColor).
			PaddingLeft(1).PaddingRight(1)
	}

	styleModal = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorCyan).
			PaddingTop(1).PaddingBottom(1)
)

func kindColor(kind string) lipgloss.Color {
	switch kind {
	case "skill":
		return colorGreen
	case "hook":
		return colorYellow
	case "model":
		return colorCyan
	case "workspace":
		return colorMagenta
	case "mcp":
		return colorBlue
	case "session":
		return colorWhite
	case "cron":
		return colorRed
	case "memory":
		return colorMagenta
	case "update":
		return colorCyan
	case "webhook":
		return colorBlue
	case "auditlog":
		return colorYellow
	}
	return colorGray
}

func kindLabel(kind string) string {
	switch kind {
	case "skill":
		return "skill"
	case "hook":
		return "hook"
	case "model":
		return "model"
	case "workspace":
		return "workspace"
	case "mcp":
		return "mcp"
	case "session":
		return "session"
	case "cron":
		return "cron"
	case "memory":
		return "memory"
	case "update":
		return "release"
	case "webhook":
		return "webhook"
	case "auditlog":
		return "audit"
	}
	return kind
}
