// Package dialog draws Crush-style modal dialogs: a rounded accent frame, a
// title followed by a ╱╱╱ gradient rule, laid over the screen behind it.
package dialog

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/adejoux/tuikit/theme"
)

// Title renders title followed by a ╱ rule that fills the remaining width
// with a gradient from DialogGradFrom to DialogGradTo.
func Title(st theme.Styles, title string, width int) string {
	rendered := st.DialogTitle.Render(title)
	if lipgloss.Width(rendered) >= width {
		return ansi.Truncate(rendered, max(0, width), "…")
	}
	n := width - lipgloss.Width(rendered) - 1
	if n <= 0 {
		return rendered
	}
	colors := lipgloss.Blend1D(n, st.DialogGradFrom, st.DialogGradTo)
	var rule strings.Builder
	for i := 0; i < n; i++ {
		style := lipgloss.NewStyle()
		if i < len(colors) && colors[i] != nil {
			style = style.Foreground(colors[i])
		}
		rule.WriteString(style.Render("╱"))
	}
	return rendered + " " + rule.String()
}

// Overlay draws dialog centred over base, which keeps the screen behind the
// dialog visible.
func Overlay(base, dialog string, width, height int) string {
	dw, dh := lipgloss.Width(dialog), lipgloss.Height(dialog)
	x := max(0, (width-dw)/2)
	y := max(0, (height-dh)/2)
	return lipgloss.NewCompositor(
		lipgloss.NewLayer(base),
		lipgloss.NewLayer(dialog).X(x).Y(y).Z(1),
	).Render()
}

// Width returns the outer width of a dialog for a screen: at least minWidth,
// at most maxWidth, and never wider than the screen.
func Width(screenWidth, minWidth, maxWidth int) int {
	return min(min(max(minWidth, screenWidth-4), maxWidth), screenWidth)
}

// Frame renders lines inside the dialog frame at the given outer width,
// cutting each line to the inner width.
func Frame(st theme.Styles, lines []string, width int) string {
	inner := InnerWidth(st, width)
	out := make([]string, len(lines))
	for i, line := range lines {
		if lipgloss.Width(line) > inner {
			line = ansi.Truncate(line, inner, "…")
		}
		out[i] = line
	}
	return st.Dialog.Width(width).Render(strings.Join(out, "\n"))
}

// InnerWidth is the content width of a dialog of the given outer width.
func InnerWidth(st theme.Styles, width int) int {
	return max(1, width-st.Dialog.GetHorizontalFrameSize())
}
