package dialog

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/adejoux/tuikit/theme"
)

func TestTitleFillsWidthWithRule(t *testing.T) {
	st := theme.NewStyles(theme.Dark())
	for _, width := range []int{5, 12, 40, 90} {
		title := Title(st, "Shortcuts", width)
		if got := lipgloss.Width(title); got > width {
			t.Fatalf("width %d: title is %d cells", width, got)
		}
		if width >= 20 && (lipgloss.Width(title) != width || !strings.Contains(ansi.Strip(title), "╱")) {
			t.Fatalf("width %d: expected a ╱ rule filling the line, got %q", width, ansi.Strip(title))
		}
	}
}

func TestFrameAndOverlayStayWithinScreen(t *testing.T) {
	st := theme.NewStyles(theme.Dark())
	for _, size := range [][2]int{{40, 10}, {80, 24}, {200, 50}} {
		w := Width(size[0], 40, 104)
		box := Frame(st, []string{Title(st, "Confirm", InnerWidth(st, w)), "", strings.Repeat("x", 300)}, w)
		if got := lipgloss.Width(box); got > size[0] || got != w {
			t.Fatalf("screen %v: dialog width %d, want %d", size, got, w)
		}
		base := strings.TrimRight(strings.Repeat(strings.Repeat(".", size[0])+"\n", size[1]), "\n")
		screen := Overlay(base, box, size[0], size[1])
		if lipgloss.Width(screen) > size[0] || lipgloss.Height(screen) > size[1] {
			t.Fatalf("screen %v: overlay is %dx%d", size, lipgloss.Width(screen), lipgloss.Height(screen))
		}
	}
}
