// Package textutil holds width-aware text helpers for terminal layouts.
package textutil

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// Ellipsis marks text that was cut.
const Ellipsis = "…"

// Wrap wraps plain text on word boundaries to width cells. Widths below 10
// return s unchanged, as wrapping would only hurt readability.
func Wrap(s string, width int) string {
	if width < 10 {
		return s
	}
	var out []string
	for _, paragraph := range strings.Split(s, "\n") {
		if lipgloss.Width(paragraph) <= width {
			out = append(out, paragraph)
			continue
		}
		line := ""
		for _, word := range strings.Fields(paragraph) {
			switch {
			case line == "":
				line = word
			case lipgloss.Width(line)+1+lipgloss.Width(word) <= width:
				line += " " + word
			default:
				out = append(out, line)
				line = word
			}
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

// CropLines returns height lines of s starting at offset. When lines remain
// below the window, the last visible line becomes an ellipsis.
func CropLines(s string, offset, height int) string {
	lines := strings.Split(s, "\n")
	if offset >= len(lines) {
		offset = max(0, len(lines)-1)
	}
	offset = max(0, offset)
	end := min(len(lines), offset+max(0, height))
	visible := append([]string(nil), lines[offset:end]...)
	if end < len(lines) && len(visible) > 0 {
		visible[len(visible)-1] = Ellipsis
	}
	return strings.Join(visible, "\n")
}

// Truncate cuts plain text to width cells, ending with an ellipsis when cut.
func Truncate(s string, width int) string {
	if width <= 0 || lipgloss.Width(s) <= width {
		return s
	}
	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes))+1 > width {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + Ellipsis
}

// TruncateANSI cuts styled text to width cells without breaking escape
// sequences.
func TruncateANSI(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	return ansi.Truncate(s, width, Ellipsis)
}

// PadRight pads s with spaces to width cells.
func PadRight(s string, width int) string {
	return s + strings.Repeat(" ", max(0, width-lipgloss.Width(s)))
}

// Plural returns word, with an s when n is not 1.
func Plural(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}
