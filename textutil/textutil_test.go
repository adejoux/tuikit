package textutil

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestWrapKeepsWordsWithinWidth(t *testing.T) {
	text := "cable sends recipes defined in YAML from a terminal user interface"
	for _, width := range []int{10, 15, 30} {
		for _, line := range strings.Split(Wrap(text, width), "\n") {
			if lipgloss.Width(line) > width {
				t.Fatalf("width %d: line %q is too wide", width, line)
			}
		}
	}
	if Wrap(text, 5) != text {
		t.Fatal("widths below 10 leave the text unchanged")
	}
	if got := Wrap("ééééé ééééé ééééé", 11); got != "ééééé ééééé\nééééé" {
		t.Fatalf("wrap must count cells, not bytes: %q", got)
	}
}

func TestCropLinesMarksHiddenLines(t *testing.T) {
	s := "a\nb\nc\nd"
	if got := CropLines(s, 1, 2); got != "b\n…" {
		t.Fatalf("CropLines = %q", got)
	}
	if got := CropLines(s, 9, 2); got != "d" {
		t.Fatalf("offset past the end shows the last line, got %q", got)
	}
	if got := CropLines(s, 0, 10); got != s {
		t.Fatalf("a tall window shows everything, got %q", got)
	}
}

func TestTruncateHelpers(t *testing.T) {
	if got := Truncate("invoices.list", 8); got != "invoice…" || lipgloss.Width(got) > 8 {
		t.Fatalf("Truncate = %q", got)
	}
	if got := Truncate("short", 8); got != "short" {
		t.Fatalf("Truncate = %q", got)
	}
	styled := lipgloss.NewStyle().Bold(true).Render("invoices.list")
	if got := TruncateANSI(styled, 8); lipgloss.Width(got) > 8 {
		t.Fatalf("TruncateANSI width = %d", lipgloss.Width(got))
	}
	if TruncateANSI(styled, 0) != "" {
		t.Fatal("zero width renders nothing")
	}
	if got := PadRight("ab", 4); got != "ab  " {
		t.Fatalf("PadRight = %q", got)
	}
	if Plural(1, "shortcut") != "shortcut" || Plural(2, "shortcut") != "shortcuts" {
		t.Fatal("Plural")
	}
}
