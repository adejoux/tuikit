package help

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/adejoux/tuikit/theme"
)

func entries() []Entry {
	return []Entry{
		{Keys: "f1 / ?", Action: "help", Description: "Show shortcuts", Group: "General"},
		{Keys: "j/k", Action: "move", Description: "Select row", Group: "Navigation", Context: "list"},
		{Keys: "enter", Action: "send", Description: "Send recipe", Group: "Actions", Context: "list", Scopes: []string{"recipes"}},
		{Keys: "ctrl+s", Action: "send draft", Description: "Send the draft", Group: "Draft", Context: "draft"},
	}
}

func key(text string) tea.KeyPressMsg {
	return tea.KeyPressMsg{Text: text, Code: []rune(text)[0]}
}

func TestFilteredByContextScopeAndQuery(t *testing.T) {
	d := New(entries(), map[string]string{"draft": "drafts"})
	if got := len(d.Filtered("list", "recipes")); got != 3 {
		t.Fatalf("list/recipes: %d entries, want 3", got)
	}
	if got := len(d.Filtered("list", "response")); got != 2 {
		t.Fatalf("scope must hide recipe-only entries: %d", got)
	}
	d.All = true
	if got := len(d.Filtered("list", "response")); got != 4 {
		t.Fatalf("all view: %d entries, want 4", got)
	}
	d.Query = "DRAFT"
	if got := d.Filtered("list", ""); len(got) != 1 || got[0].Keys != "ctrl+s" {
		t.Fatalf("query must match case-insensitively: %#v", got)
	}
}

func TestUpdateTypesTogglesAndCloses(t *testing.T) {
	d := New(entries(), nil)
	for _, r := range "send" {
		if d.Update(key(string(r)), 24) {
			t.Fatal("typing must not close")
		}
	}
	if d.Query != "send" {
		t.Fatalf("query = %q", d.Query)
	}
	d.Update(tea.KeyPressMsg{Code: tea.KeyBackspace}, 24)
	if d.Query != "sen" {
		t.Fatalf("backspace: query = %q", d.Query)
	}
	d.Update(tea.KeyPressMsg{Code: tea.KeyTab}, 24)
	if !d.All {
		t.Fatal("tab must switch to the all view")
	}
	if !d.Update(tea.KeyPressMsg{Code: tea.KeyEscape}, 24) {
		t.Fatal("esc must close")
	}
}

func TestViewFitsTheScreen(t *testing.T) {
	st := theme.NewStyles(theme.Dark())
	many := entries()
	for i := 0; i < 60; i++ {
		many = append(many, Entry{Keys: "x", Action: "action", Description: "A long description of an action", Group: "Group " + string(rune('A'+i%6))})
	}
	for _, size := range [][2]int{{60, 20}, {80, 24}, {120, 36}, {200, 50}} {
		for _, all := range []bool{false, true} {
			d := New(many, nil)
			d.All = all
			d.Offset = 1000
			view := d.View(st, "list", "recipes", size[0], size[1])
			if w, h := lipgloss.Width(view), lipgloss.Height(view); w > size[0] || h > size[1] {
				t.Fatalf("screen %v all=%v: dialog is %dx%d", size, all, w, h)
			}
			plain := ansi.Strip(view)
			if !strings.Contains(plain, "Shortcuts") || !strings.Contains(plain, "esc close") {
				t.Fatalf("screen %v: missing title or close hint:\n%s", size, plain)
			}
		}
	}
}
