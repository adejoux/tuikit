package form

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/adejoux/tuikit/keyhints"
	"github.com/adejoux/tuikit/theme"
)

func press(s string) tea.KeyPressMsg {
	switch s {
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case "shift+tab":
		return tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "left":
		return tea.KeyPressMsg{Code: tea.KeyLeft}
	case "right":
		return tea.KeyPressMsg{Code: tea.KeyRight}
	case "ctrl+s":
		return tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}
	case "ctrl+n":
		return tea.KeyPressMsg{Code: 'n', Mod: tea.ModCtrl}
	}
	return tea.KeyPressMsg{Code: []rune(s)[0], Text: s}
}

func typeText(f *Form, text string) {
	for _, r := range text {
		f.Update(press(string(r)))
	}
}

func TestCompletionUsesCurrentCommaSeparatedItem(t *testing.T) {
	f, _ := New("Labels", CompletionField("labels", "Labels", "", Strings([]string{"backend", "security", "security-review", "frontend"}), false))
	field := f.Field("labels")
	field.Input.SetValue("backend, sec")
	field.RefreshSuggestions()
	if len(field.Suggestions) != 2 || field.Suggestions[0].Value != "security" {
		t.Fatalf("suggestions = %#v", field.Suggestions)
	}
	if !field.AcceptSuggestion() || field.Input.Value() != "backend, security" {
		t.Fatalf("value = %q", field.Input.Value())
	}
}

func TestCompletionSkipsUsedItemsAndKeepsTypedCase(t *testing.T) {
	f, _ := New("Labels", CompletionField("labels", "Labels", "", Strings([]string{"api", "Backend"}), false))
	field := f.Field("labels")
	field.Input.SetValue("Backend, back")
	field.RefreshSuggestions()
	if len(field.Suggestions) != 0 {
		t.Fatalf("an item already entered must not be suggested: %#v", field.Suggestions)
	}
	field.Input.SetValue("Backend, API")
	field.RefreshSuggestions()
	if field.AcceptSuggestion() || field.Input.Value() != "Backend, API" {
		t.Fatalf("a value typed in full keeps its case, got %q", field.Input.Value())
	}
	field.Input.SetValue("back")
	field.RefreshSuggestions()
	if !field.AcceptSuggestion() || field.Input.Value() != "Backend" {
		t.Fatalf("completing a prefix uses the option as written, got %q", field.Input.Value())
	}
}

func TestWholeCompletionMatchesDescriptionAndStoresValue(t *testing.T) {
	people := []Option{{Label: "John Smith", Value: "jsmith", Description: "@jsmith"}, {Label: "Alice Martin", Value: "amartin"}}
	f, _ := New("Assign", CompletionField("assignee", "Assignee", "", people, true))
	field := f.Field("assignee")
	typeText(f, "John")
	if len(field.Suggestions) == 0 || field.Suggestions[0].Label != "John Smith" {
		t.Fatalf("suggestions = %#v", field.Suggestions)
	}
	f.Update(press("tab"))
	if got := f.Value("assignee"); got != "jsmith" {
		t.Fatalf("value = %q, want the option value", got)
	}
}

func TestPickerRanksAndSelects(t *testing.T) {
	options := []Option{
		{Label: "A product labels and templates", Value: "scattered"},
		{Label: "Z platform API", Value: "consecutive"},
		{Label: "Billing", Value: "epc_bill", SearchText: "epc_bill billing"},
	}
	f, _ := New("Pick", PickerField("project", "Project", options, "", false))
	f.Update(press("enter"))
	if f.Picker == nil {
		t.Fatal("enter on a picker field must open the picker")
	}
	typeText(f, "plat")
	if len(f.Picker.Filtered) != 2 || f.Picker.Filtered[0] != 1 {
		t.Fatalf("picker order = %#v", f.Picker.Filtered)
	}
	f.Update(press("ctrl+n"))
	f.Update(press("enter"))
	if f.Picker != nil || f.Value("project") != "scattered" {
		t.Fatalf("picker must close on the chosen option, got %q", f.Value("project"))
	}
	if got := Rank(options, "epc_bill"); len(got) != 1 || got[0] != 2 {
		t.Fatalf("search text must match: %#v", got)
	}
}

func TestNavigationSelectAndSubmit(t *testing.T) {
	f, _ := New("New recipe",
		TextField("id", "Id", "invoices.list", "list"),
		SelectField("method", "Method", Strings([]string{"GET", "POST", "PUT"}), "POST", true),
		TextField("url", "URL", "", ""),
	)
	if !f.Fields[0].Input.Focused() {
		t.Fatal("the first text field must be focused")
	}
	f.Update(press("tab"))
	f.Update(press("right"))
	if f.Value("method") != "PUT" {
		t.Fatalf("right must select the next option, got %q", f.Value("method"))
	}
	f.Update(press("left"))
	f.Update(press("left"))
	if f.Value("method") != "GET" {
		t.Fatalf("left must select the previous option, got %q", f.Value("method"))
	}
	f.Update(press("down"))
	if f.Cursor != 2 || !f.Fields[2].Input.Focused() || f.Fields[0].Input.Focused() {
		t.Fatal("down on a select moves to the next field and moves the focus")
	}
	if got := f.Missing(); len(got) != 0 {
		t.Fatalf("missing = %v", got)
	}
	f.Field("method").Required, f.Field("url").Required = true, true
	if got := f.Missing(); len(got) != 1 || got[0] != "URL" {
		t.Fatalf("missing = %v", got)
	}
	typeText(f, "https://x")
	if r, _ := f.Update(press("enter")); r != Submitted {
		t.Fatal("enter on the last field submits")
	}
	if r, _ := f.Update(press("ctrl+s")); r != Submitted {
		t.Fatal("ctrl+s submits")
	}
	if r, _ := f.Update(press("esc")); r != Cancelled {
		t.Fatal("esc cancels")
	}
	f.Update(press("shift+tab"))
	if f.Cursor != 1 {
		t.Fatalf("shift+tab moves back, cursor = %d", f.Cursor)
	}
	f.SetValue("method", "POST")
	f.SetValue("id", "create")
	if f.Value("method") != "POST" || f.Value("id") != "create" {
		t.Fatal("SetValue")
	}
}

func TestTextFieldsLeaveCtrlNFree(t *testing.T) {
	f, _ := New("x", TextField("a", "A", "", "abc"))
	f.Update(press("ctrl+n"))
	if f.Value("a") != "abc" || f.Cursor != 0 {
		t.Fatal("ctrl+n must not act inside a text field")
	}
}

func TestViewFitsWidth(t *testing.T) {
	st := theme.NewStyles(theme.Dark())
	f, _ := New("New recipe",
		TextField("id", "Identifier with a very long label", "", strings.Repeat("x", 200)),
		SelectField("method", "Method", Strings([]string{"GET", "POST"}), "GET", true),
		PickerField("folder", "Folder", Strings([]string{"auth", "invoices"}), "auth", false),
		CompletionField("tags", "Tags", "", Strings([]string{"alpha", "beta"}), false),
	)
	f.Cursor = 3
	f.Field("tags").Input.SetValue("a")
	f.Field("tags").RefreshSuggestions()
	for _, width := range []int{40, 80, 140} {
		view := f.View(st, width, 30, keyhints.Hint{Key: "ctrl+g", Desc: "assist"})
		if w := lipgloss.Width(view); w > width {
			t.Fatalf("width %d: view is %d cells", width, w)
		}
		if !strings.Contains(ansi.Strip(view), "esc cancel") {
			t.Fatalf("width %d: cancel hint missing", width)
		}
	}
	f.Cursor = 2
	f.OpenPicker()
	if !strings.Contains(ansi.Strip(f.View(st, 80, 20)), "2 / 2 matches") {
		t.Fatal("picker view must show the match count")
	}
}
