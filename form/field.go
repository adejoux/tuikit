// Package form is a keyboard-driven form: text fields, completion fields,
// inline selects and searchable pickers, laid out as label/value rows.
//
// The application keeps its own keys (an assistant, a synthesis action…) and
// passes every other message to Form.Update, which reports when the form is
// submitted or cancelled.
package form

import (
	"sort"
	"strings"

	"charm.land/bubbles/v2/textinput"

	"github.com/adejoux/tuikit/fuzzy"
)

// Kind is the kind of a field.
type Kind int

const (
	// Text is a free text field.
	Text Kind = iota
	// Completion is a text field with ranked suggestions from Candidates.
	Completion
	// Select cycles through a few Options with ←/→ or space.
	Select
	// Picker opens a searchable list of Options on enter.
	Picker
)

// Option is one choice of a select, picker or completion field.
type Option struct {
	Label       string
	Value       string
	Description string
	// SearchText adds words that match the option without being shown,
	// such as an ID or aliases.
	SearchText string
}

// Field is one row of a form.
type Field struct {
	Key      string
	Label    string
	Kind     Kind
	Hint     string // placeholder of text and completion fields
	Required bool

	// Value is the initial value of text and completion fields.
	Value string
	Input textinput.Model

	// Options and Selected back select and picker fields.
	Options  []Option
	Selected int

	// Candidates are the completion choices. By default a completion field
	// holds a comma-separated list and completes its last item; Whole
	// completes the entire value instead.
	Candidates  []Option
	Whole       bool
	Suggestions []Option
	SuggestPos  int
}

// MaxSuggestions is the number of completion suggestions shown.
const MaxSuggestions = 6

// TextField returns a free text field.
func TextField(key, label, hint, value string) Field {
	return Field{Key: key, Label: label, Kind: Text, Hint: hint, Value: value}
}

// CompletionField returns a text field completed from candidates. whole
// completes the entire value instead of the last comma-separated item.
func CompletionField(key, label, hint string, candidates []Option, whole bool) Field {
	return Field{Key: key, Label: label, Kind: Completion, Hint: hint, Candidates: candidates, Whole: whole}
}

// SelectField returns a select field with value selected, or the first
// option when value is not found.
func SelectField(key, label string, options []Option, value string, required bool) Field {
	return choiceField(Select, key, label, options, value, required)
}

// PickerField returns a picker field with value selected, or the first
// option when value is not found.
func PickerField(key, label string, options []Option, value string, required bool) Field {
	return choiceField(Picker, key, label, options, value, required)
}

func choiceField(kind Kind, key, label string, options []Option, value string, required bool) Field {
	f := Field{Key: key, Label: label, Kind: kind, Options: options, Required: required}
	for i, o := range options {
		if o.Value == value {
			f.Selected = i
			break
		}
	}
	return f
}

// Strings returns one option per value, used as label, value and search
// text.
func Strings(values []string) []Option {
	out := make([]Option, 0, len(values))
	for _, v := range values {
		out = append(out, Option{Label: v, Value: v, SearchText: v})
	}
	return out
}

// YesNo returns the two options of a boolean select, "false" first.
func YesNo(noLabel, yesLabel string) []Option {
	return []Option{{Label: noLabel, Value: "false"}, {Label: yesLabel, Value: "true"}}
}

// IsText reports whether the field is edited through a text input.
func (f *Field) IsText() bool {
	return f.Kind == Text || f.Kind == Completion
}

// CurrentValue returns the field value: the trimmed text of text fields, the
// selected option value of select and picker fields.
func (f *Field) CurrentValue() string {
	if f.IsText() {
		return strings.TrimSpace(f.Input.Value())
	}
	if f.Selected < 0 || f.Selected >= len(f.Options) {
		return ""
	}
	return f.Options[f.Selected].Value
}

// Rank returns the indexes of the options matching query, best first. An
// empty query keeps every option in its original order.
func Rank(options []Option, query string) []int {
	query = strings.ToLower(strings.TrimSpace(query))
	type ranked struct{ index, score int }
	matches := make([]ranked, 0, len(options))
	for i, o := range options {
		score, ok := score(o, query)
		if ok {
			matches = append(matches, ranked{i, score})
		}
	}
	if query != "" {
		sort.SliceStable(matches, func(i, j int) bool { return matches[i].score > matches[j].score })
	}
	out := make([]int, len(matches))
	for i, m := range matches {
		out[i] = m.index
	}
	return out
}

func score(o Option, query string) (int, bool) {
	search := strings.TrimSpace(o.SearchText)
	if search == "" {
		search = o.Label + " " + o.Value
	}
	return fuzzy.ScoreFields(query,
		fuzzy.Field{Value: o.Label, Weight: 3},
		fuzzy.Field{Value: o.Value, Weight: 2},
		fuzzy.Field{Value: o.Description, Weight: 1},
		fuzzy.Field{Value: search, Weight: 1},
	)
}

// RefreshSuggestions recomputes the suggestions of a completion field from
// its current text.
func (f *Field) RefreshSuggestions() {
	if f.Kind != Completion {
		return
	}
	value := f.Input.Value()
	token := f.token(value)
	f.Suggestions = f.Suggestions[:0]
	if token == "" {
		f.SuggestPos = 0
		return
	}
	used := map[string]bool{}
	if !f.Whole {
		parts := strings.Split(value, ",")
		for _, part := range parts[:len(parts)-1] {
			if part = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(part), "@")); part != "" {
				used[strings.ToLower(part)] = true
			}
		}
	}
	type ranked struct {
		option Option
		score  int
	}
	var matches []ranked
	for _, c := range f.Candidates {
		if used[strings.ToLower(c.Value)] {
			continue
		}
		if s, ok := score(c, strings.ToLower(token)); ok {
			matches = append(matches, ranked{c, s})
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score > matches[j].score
		}
		return strings.ToLower(matches[i].option.Label) < strings.ToLower(matches[j].option.Label)
	})
	for i := 0; i < len(matches) && i < MaxSuggestions; i++ {
		f.Suggestions = append(f.Suggestions, matches[i].option)
	}
	if f.SuggestPos >= len(f.Suggestions) {
		f.SuggestPos = 0
	}
}

// AcceptSuggestion replaces the item being typed with the selected
// suggestion. It returns false when there is nothing to accept, including
// when the item is already typed in full with another case: the value is
// then kept as typed.
func (f *Field) AcceptSuggestion() bool {
	if len(f.Suggestions) == 0 {
		return false
	}
	selected := f.Suggestions[f.SuggestPos].Value
	value := f.Input.Value()
	token := f.token(value)
	if token != selected && strings.EqualFold(token, selected) {
		f.Suggestions, f.SuggestPos = nil, 0
		return false
	}
	prefix := ""
	if !f.Whole {
		if i := strings.LastIndex(value, ","); i >= 0 {
			if prefix = strings.TrimSpace(value[:i+1]); prefix != "" {
				prefix += " "
			}
		}
	}
	f.Input.SetValue(prefix + selected)
	f.Input.CursorEnd()
	f.Value = f.Input.Value()
	f.Suggestions, f.SuggestPos = nil, 0
	return true
}

// token is the part of value being completed.
func (f *Field) token(value string) string {
	if !f.Whole {
		if i := strings.LastIndex(value, ","); i >= 0 {
			value = value[i+1:]
		}
	}
	return strings.TrimSpace(value)
}
