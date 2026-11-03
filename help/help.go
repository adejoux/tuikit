// Package help is the searchable shortcut dialog opened with f1 or ?. It
// shows the shortcuts of the current context by default and every shortcut
// on tab, grouped in balanced columns, after Crush's dialogs.
package help

import (
	"fmt"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/adejoux/tuikit/dialog"
	"github.com/adejoux/tuikit/keyhints"
	"github.com/adejoux/tuikit/textutil"
	"github.com/adejoux/tuikit/theme"
)

// Entry is one shortcut.
type Entry struct {
	Keys        string
	Action      string
	Description string
	Group       string
	// Context limits the entry to one help context, such as "list", "form"
	// or "draft". Empty means every context.
	Context string
	// Scopes limits the entry to some screens or panels of its context.
	// Empty means every scope.
	Scopes []string
}

// Dialog is the state of an open shortcut dialog.
type Dialog struct {
	Entries []Entry
	// ContextLabels names contexts in the "all" view, e.g. "form": "forms".
	// A context without a label adds nothing to its group names.
	ContextLabels map[string]string

	Query  string
	All    bool
	Offset int
}

// New returns a dialog over entries.
func New(entries []Entry, contextLabels map[string]string) *Dialog {
	return &Dialog{Entries: entries, ContextLabels: contextLabels}
}

// Filtered returns the entries shown for context and scope, matching the
// query. In the "all" view, context and scope are ignored.
func (d *Dialog) Filtered(context, scope string) []Entry {
	query := strings.ToLower(strings.TrimSpace(d.Query))
	out := make([]Entry, 0, len(d.Entries))
	for _, e := range d.Entries {
		if !d.All && e.Context != "" && e.Context != context {
			continue
		}
		if !d.All && len(e.Scopes) > 0 && !contains(e.Scopes, scope) {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(e.Keys+" "+e.Action+" "+e.Description+" "+e.Group), query) {
			continue
		}
		out = append(out, e)
	}
	return out
}

// Update handles a key press. It returns true when the dialog must close.
// height is the screen height, used for page scrolling.
func (d *Dialog) Update(msg tea.KeyPressMsg, height int) (closed bool) {
	switch msg.String() {
	case "esc", "f1":
		return true
	case "tab":
		d.All = !d.All
		d.Offset = 0
	case "up":
		d.Offset = max(0, d.Offset-1)
	case "down":
		d.Offset++
	case "pgdown":
		d.Offset += max(1, height/2)
	case "pgup":
		d.Offset = max(0, d.Offset-max(1, height/2))
	case "backspace":
		if len(d.Query) > 0 {
			_, size := utf8.DecodeLastRuneInString(d.Query)
			d.Query = d.Query[:len(d.Query)-size]
			d.Offset = 0
		}
	default:
		if text := msg.Text; text != "" && utf8.RuneCountInString(text) == 1 && text >= " " {
			d.Query += text
			d.Offset = 0
		}
	}
	return false
}

// View renders the dialog for a screen of the given size. Lay it over the
// screen with dialog.Overlay.
func (d *Dialog) View(st theme.Styles, context, scope string, screenWidth, screenHeight int) string {
	width := dialog.Width(screenWidth, 40, 104)
	inner := dialog.InnerWidth(st, width)

	here, all := st.DialogPillOn, st.DialogPill
	if d.All {
		here, all = all, here
	}
	toggle := here.Render("this screen") + all.Render("all")
	title := dialog.Title(st, "Shortcuts", inner-lipgloss.Width(toggle)-1) + " " + toggle

	prompt := st.DialogPrompt.Render("> ")
	if d.Query == "" {
		prompt += st.Muted.Render("type to filter")
	} else {
		prompt += d.Query + st.DialogPrompt.Render("▏")
	}

	entries := d.Filtered(context, scope)
	body := d.columns(st, entries, inner)
	// Esc is pinned: the way out must always be visible.
	closeHint := keyhints.Hint{Key: "esc", Desc: "close"}
	count := keyhints.Hint{Key: fmt.Sprint(len(entries)), Desc: textutil.Plural(len(entries), "shortcut")}

	// Fixed parts: title, blank, prompt, blank, (body), blank, footer, plus
	// a line above and below so the dialog visibly floats.
	maxBody := max(1, screenHeight-2-st.Dialog.GetVerticalFrameSize()-6)
	d.Offset = min(d.Offset, max(0, len(body)-maxBody))
	visible := body[d.Offset:min(len(body), d.Offset+maxBody)]
	scroll := keyhints.Hint{Key: "↑↓", Desc: "scroll"}
	if len(body) > maxBody {
		scroll.Desc = fmt.Sprintf("scroll %d/%d", d.Offset+len(visible), len(body))
	}
	footer := keyhints.Render(st, []keyhints.Hint{{Key: "tab", Desc: "this screen / all"}, scroll, count}, &closeHint, inner)

	lines := []string{title, "", prompt, ""}
	lines = append(lines, visible...)
	lines = append(lines, "", footer)
	return dialog.Frame(st, lines, width)
}

// columns lays groups out in as many columns as the width allows, each group
// a heading followed by aligned "key  action" rows. Narrow screens and
// searches get one column with the descriptions.
func (d *Dialog) columns(st theme.Styles, entries []Entry, width int) []string {
	if len(entries) == 0 {
		return []string{st.Muted.Render("No matching shortcuts")}
	}
	type group struct {
		name    string
		entries []Entry
	}
	var groups []*group
	index := map[string]*group{}
	for _, e := range entries {
		name := e.Group
		if d.All {
			name = d.groupLabel(e.Group, e.Context)
		}
		g, ok := index[name]
		if !ok {
			g = &group{name: name}
			index[name] = g
			groups = append(groups, g)
		}
		g.entries = append(g.entries, e)
	}

	keyWidth, actionWidth := 0, 0
	for _, e := range entries {
		keyWidth = max(keyWidth, lipgloss.Width(e.Keys))
		actionWidth = max(actionWidth, lipgloss.Width(e.Action))
	}
	keyWidth = min(keyWidth, 18)
	const gap = 4
	columnWidth := keyWidth + 2 + actionWidth
	columns := max(1, min(3, (width+gap)/(columnWidth+gap)))
	if strings.TrimSpace(d.Query) != "" {
		columns = 1
	}
	withDescription := columns == 1

	block := func(g *group) []string {
		lines := []string{st.Heading.Render(g.name)}
		for _, e := range g.entries {
			k := st.HelpKey.Render(textutil.PadRight(textutil.Truncate(e.Keys, keyWidth), keyWidth))
			line := k + "  " + st.HelpAction.Render(e.Action)
			if withDescription {
				line = k + "  " + st.HelpAction.Render(textutil.PadRight(e.Action, actionWidth)) + "  " + st.HelpDesc.Render(e.Description)
			}
			lines = append(lines, line)
		}
		return lines
	}
	if columns == 1 {
		var out []string
		for i, g := range groups {
			if i > 0 {
				out = append(out, "")
			}
			out = append(out, block(g)...)
		}
		return out
	}

	// Place each group in the currently shortest column, which keeps the
	// columns balanced while preserving the group order down each column.
	cols := make([][]string, columns)
	for _, g := range groups {
		shortest := 0
		for i := range cols {
			if len(cols[i]) < len(cols[shortest]) {
				shortest = i
			}
		}
		if len(cols[shortest]) > 0 {
			cols[shortest] = append(cols[shortest], "")
		}
		cols[shortest] = append(cols[shortest], block(g)...)
	}
	rendered := make([]string, 0, columns*2)
	for i, col := range cols {
		if i > 0 {
			rendered = append(rendered, strings.Repeat(" ", gap))
		}
		rendered = append(rendered, lipgloss.NewStyle().Width(columnWidth).Render(strings.Join(col, "\n")))
	}
	return strings.Split(lipgloss.JoinHorizontal(lipgloss.Top, rendered...), "\n")
}

// groupLabel adds the context to a group name in the "all" view, unless the
// group already names it.
func (d *Dialog) groupLabel(group, context string) string {
	label, ok := d.ContextLabels[context]
	if !ok || strings.HasPrefix(strings.ToLower(label), strings.ToLower(group)) ||
		strings.HasPrefix(strings.ToLower(group), strings.ToLower(context)) {
		return group
	}
	return group + " · " + label
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
