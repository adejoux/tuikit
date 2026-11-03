package form

import (
	"fmt"
	"strings"

	"github.com/adejoux/tuikit/keyhints"
	"github.com/adejoux/tuikit/textutil"
	"github.com/adejoux/tuikit/theme"
)

// labelWidth is the width of the label column.
const labelWidth = 18

// View renders the form, or its open picker, in width × height cells.
// extra hints are shown before "esc cancel", e.g. an application key such
// as {"ctrl+g", "assist"}.
func (f *Form) View(st theme.Styles, width, height int, extra ...keyhints.Hint) string {
	if f.Picker != nil {
		return f.pickerView(st, width, height)
	}
	contentW := max(28, min(88, width-8))
	lines := []string{st.Title.Render(f.Title)}
	if f.Context != "" {
		lines = append(lines, st.Muted.Render(f.Context))
	}
	lines = append(lines, "")
	for i := range f.Fields {
		field := &f.Fields[i]
		label := field.Label
		if field.Required {
			label += " *"
		}
		value := ""
		switch field.Kind {
		case Text, Completion:
			field.Input.SetWidth(max(18, contentW-labelWidth-6))
			value = field.Input.View()
		case Select:
			if len(field.Options) == 0 {
				value = st.Muted.Render("(no options available)")
			} else {
				value = "‹ " + field.Options[field.Selected].Label + " ›  " + st.Muted.Render("←/→")
			}
		case Picker:
			if len(field.Options) == 0 {
				value = st.Muted.Render("(no options available)")
			} else {
				value = field.Options[field.Selected].Label + "  " + st.Muted.Render("enter to choose")
			}
		}
		line := textutil.PadRight(textutil.Truncate(label, labelWidth), labelWidth) + " " + value
		if i == f.Cursor {
			lines = append(lines, st.SelectedLine(line, contentW))
		} else {
			lines = append(lines, st.Line(line, contentW))
		}
		if i == f.Cursor && field.Kind == Completion && len(field.Suggestions) > 0 {
			for j, s := range field.Suggestions {
				prefix := "    "
				if j == field.SuggestPos {
					prefix = "  › "
				}
				text := s.Label
				if s.Description != "" {
					text += "  " + s.Description
				}
				lines = append(lines, st.Muted.Render(textutil.Truncate(prefix+text, contentW)))
			}
			lines = append(lines, keyhints.Pairs(st, contentW, "↑/↓", "select", "tab/enter", "complete"))
		}
	}
	submit := f.SubmitLabel
	if submit == "" {
		submit = "save"
	}
	hints := []keyhints.Hint{{Key: "tab", Desc: "next"}, {Key: "shift+tab", Desc: "previous"}, {Key: "enter", Desc: "choose/next"}, {Key: "ctrl+s", Desc: submit}}
	hints = append(hints, extra...)
	cancel := keyhints.Hint{Key: "esc", Desc: "cancel"}
	lines = append(lines, "", keyhints.Render(st, hints, &cancel, width))
	return textutil.CropLines(strings.Join(lines, "\n"), 0, height)
}

func (f *Form) pickerView(st theme.Styles, width, height int) string {
	ps := f.Picker
	field := &f.Fields[ps.FieldIndex]
	contentW := max(28, min(88, width-8))
	ps.Query.SetWidth(max(20, contentW-12))
	lines := []string{
		st.Title.Render("Select " + strings.ToLower(field.Label)),
		st.Muted.Render("Type to filter by name, ID or description."),
		"",
		st.DialogPrompt.Render("> ") + ps.Query.View(),
		"",
	}
	maxRows := max(3, height-9)
	start := 0
	if ps.Cursor >= maxRows {
		start = ps.Cursor - maxRows + 1
	}
	end := min(len(ps.Filtered), start+maxRows)
	for i := start; i < end; i++ {
		option := field.Options[ps.Filtered[i]]
		line := option.Label
		if option.Description != "" {
			line += "  " + st.Muted.Render(option.Description)
		}
		if i == ps.Cursor {
			lines = append(lines, st.SelectedLine(line, contentW))
		} else {
			lines = append(lines, st.Line(line, contentW))
		}
	}
	if len(ps.Filtered) == 0 {
		lines = append(lines, st.Muted.Render("  No matching options"))
	}
	lines = append(lines, "",
		st.Muted.Render(fmt.Sprintf("%d / %d matches", len(ps.Filtered), len(field.Options))),
		keyhints.Pairs(st, width, "type", "filter", "↑/↓", "navigate", "enter", "select", "esc", "cancel"),
	)
	return textutil.CropLines(strings.Join(lines, "\n"), 0, height)
}
