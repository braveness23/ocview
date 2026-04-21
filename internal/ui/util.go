package ui

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func shortPath(p string) string {
	home, _ := os.UserHomeDir()
	return strings.Replace(p, home, "~", 1)
}

func fmtTimestamp(ms int64) string {
	return time.UnixMilli(ms).Format("Jan 2, 2006 03:04 PM")
}

func fmtTimeISO(iso string) string {
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return iso
	}
	return t.Format("Jan 2, 2006 03:04 PM")
}

func fmtCost(n float64) string {
	if n == 0 {
		return "free"
	}
	return fmt.Sprintf("$%.4f/M", n)
}

func fmtNumber(n int) string {
	if n == 0 {
		return "0"
	}
	s := fmt.Sprintf("%d", n)
	result := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}
	return result
}

// wrapText splits text on newlines then word-wraps each line to width.
func wrapText(text string, width int) []string {
	if width <= 0 {
		return strings.Split(text, "\n")
	}
	var result []string
	for _, rawLine := range strings.Split(text, "\n") {
		if rawLine == "" {
			result = append(result, "")
			continue
		}
		pos := 0
		for pos < len(rawLine) {
			if pos+width >= len(rawLine) {
				result = append(result, rawLine[pos:])
				break
			}
			end := pos + width
			spaceIdx := strings.LastIndex(rawLine[pos:end], " ")
			if spaceIdx > 0 {
				result = append(result, rawLine[pos:pos+spaceIdx])
				pos = pos + spaceIdx + 1
			} else {
				result = append(result, rawLine[pos:end])
				pos = end
			}
		}
	}
	return result
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	return s + strings.Repeat(" ", n-len(s))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n > 2 {
		return s[:n-2] + ".."
	}
	return s[:n]
}

func divider(width int) string {
	if width <= 0 {
		return ""
	}
	return strings.Repeat("─", width)
}

func mdLineStyle(line string) (color, bold string) {
	switch {
	case strings.HasPrefix(line, "# "):
		return "cyan", "bold"
	case strings.HasPrefix(line, "## "):
		return "yellow", "bold"
	case strings.HasPrefix(line, "### ") || strings.HasPrefix(line, "#### "):
		return "green", ""
	case strings.HasPrefix(line, "```"):
		return "gray", ""
	case strings.TrimSpace(line) == "---":
		return "gray", ""
	default:
		return "white", ""
	}
}
