package keyhints

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/adejoux/tuikit/theme"
)

func TestRenderFitsWidthAndKeepsPinnedHint(t *testing.T) {
	st := theme.NewStyles(theme.Dark())
	hints := []Hint{{"j/k", "navigate"}, {"enter", "details"}, {"n", "new"}, {"i", "browse import"}, {"e", "edit"}, {"/", "search"}}
	pinned := More
	for width := 6; width <= 120; width++ {
		line := Render(st, hints, &pinned, width)
		plain := ansi.Strip(line)
		if got := lipgloss.Width(line); got > width {
			t.Fatalf("width %d: rendered %d cells: %q", width, got, plain)
		}
		if !strings.HasSuffix(plain, "? more") {
			t.Fatalf("width %d: pinned hint missing: %q", width, plain)
		}
		if strings.Contains(plain, "•  •") || strings.HasPrefix(plain, " •") {
			t.Fatalf("width %d: dangling separator: %q", width, plain)
		}
	}
	full := ansi.Strip(Render(st, hints, &pinned, 200))
	if want := "j/k navigate • enter details • n new • i browse import • e edit • / search • ? more"; full != want {
		t.Fatalf("full line:\n got %q\nwant %q", full, want)
	}
	if cut := ansi.Strip(Render(st, hints, &pinned, 40)); !strings.Contains(cut, "…") {
		t.Fatalf("a cut line must show an ellipsis: %q", cut)
	}
}

func TestRenderWithoutPinnedAndZeroWidth(t *testing.T) {
	st := theme.NewStyles(theme.Light())
	if got := Render(st, []Hint{{"q", "quit"}}, nil, 0); got != "" {
		t.Fatalf("zero width must render nothing, got %q", got)
	}
	if got := ansi.Strip(Pairs(st, 0, "a", "add", "d", "delete")); got != "a add • d delete" {
		t.Fatalf("pairs = %q", got)
	}
}
