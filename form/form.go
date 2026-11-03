package form

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/adejoux/tuikit/fields"
)

// Result tells the application what a message did to the form.
type Result int

const (
	// Continue means the form stays open.
	Continue Result = iota
	// Submitted means the user asked to submit (ctrl+s, or enter on the last
	// field). The form stays as it is; the application validates and closes
	// it.
	Submitted
	// Cancelled means the user left the form with esc.
	Cancelled
)

// Form is an ordered set of fields with a cursor, and the picker of a picker
// field while it is open.
type Form struct {
	Title   string
	Context string // muted line under the title
	// SubmitLabel names the submit action in the key hints: "create",
	// "save", "import"… Defaults to "save".
	SubmitLabel string

	Fields []Field
	Cursor int
	Picker *PickerState
}

// PickerState is an open picker: the field it fills, its query and the
// matching option indexes.
type PickerState struct {
	FieldIndex int
	Query      textinput.Model
	Filtered   []int
	Cursor     int
}

// New returns a form over fields, with text inputs set up and the first
// field focused. Run the returned command to start the cursor blink.
func New(title string, fs ...Field) (*Form, tea.Cmd) {
	f := &Form{Title: title, Fields: fs}
	for i := range f.Fields {
		field := &f.Fields[i]
		if !field.IsText() {
			continue
		}
		in := fields.NewTextInput()
		in.Prompt = ""
		in.Placeholder = field.Hint
		in.CharLimit = 0
		in.SetWidth(48)
		in.SetValue(field.Value)
		field.Input = in
	}
	return f, f.focus()
}

// Field returns the field with key, or nil.
func (f *Form) Field(key string) *Field {
	for i := range f.Fields {
		if f.Fields[i].Key == key {
			return &f.Fields[i]
		}
	}
	return nil
}

// Value returns the value of the field with key, or "".
func (f *Form) Value(key string) string {
	if field := f.Field(key); field != nil {
		return field.CurrentValue()
	}
	return ""
}

// Bool reports whether the field with key holds "true".
func (f *Form) Bool(key string) bool {
	return strings.EqualFold(f.Value(key), "true")
}

// SetValue sets a text field's text, or selects the option with that value.
func (f *Form) SetValue(key, value string) {
	field := f.Field(key)
	if field == nil {
		return
	}
	if field.IsText() {
		field.Value = value
		field.Input.SetValue(value)
		return
	}
	for i, o := range field.Options {
		if o.Value == value {
			field.Selected = i
			return
		}
	}
}

// Missing returns the labels of required fields left empty.
func (f *Form) Missing() []string {
	var out []string
	for i := range f.Fields {
		if f.Fields[i].Required && f.Fields[i].CurrentValue() == "" {
			out = append(out, f.Fields[i].Label)
		}
	}
	return out
}

// Move moves the cursor by delta fields, wrapping around.
func (f *Form) Move(delta int) tea.Cmd {
	if len(f.Fields) == 0 {
		return nil
	}
	f.Cursor = (f.Cursor + delta%len(f.Fields) + len(f.Fields)) % len(f.Fields)
	return f.focus()
}

// Blur removes the focus from every text input, e.g. while the application
// shows a dialog over the form.
func (f *Form) Blur() {
	for i := range f.Fields {
		if f.Fields[i].IsText() {
			f.Fields[i].Input.Blur()
		}
	}
}

// Focus gives the focus back to the field under the cursor.
func (f *Form) Focus() tea.Cmd { return f.focus() }

func (f *Form) focus() tea.Cmd {
	var cmds []tea.Cmd
	for i := range f.Fields {
		field := &f.Fields[i]
		if !field.IsText() {
			continue
		}
		if i == f.Cursor && f.Picker == nil {
			cmds = append(cmds, field.Input.Focus())
		} else {
			field.Input.Blur()
		}
	}
	return tea.Batch(cmds...)
}

// Update handles a message. Key presses drive the form; other messages go to
// the focused input.
func (f *Form) Update(msg tea.Msg) (Result, tea.Cmd) {
	if len(f.Fields) == 0 {
		return Continue, nil
	}
	if f.Picker != nil {
		return Continue, f.updatePicker(msg)
	}
	field := &f.Fields[f.Cursor]
	key, isKey := msg.(tea.KeyPressMsg)
	if isKey {
		switch key.String() {
		case "esc":
			return Cancelled, nil
		case "ctrl+s":
			return Submitted, nil
		case "tab":
			if field.Kind == Completion && field.AcceptSuggestion() {
				return Continue, nil
			}
			return Continue, f.Move(1)
		case "shift+tab":
			return Continue, f.Move(-1)
		case "up", "down":
			delta := 1
			if key.String() == "up" {
				delta = -1
			}
			if field.Kind == Completion && len(field.Suggestions) > 0 {
				n := len(field.Suggestions)
				field.SuggestPos = (field.SuggestPos + delta + n) % n
				return Continue, nil
			}
			if !field.IsText() {
				return Continue, f.Move(delta)
			}
		case "left", "right", "space":
			if field.Kind == Select && len(field.Options) > 0 {
				delta := 1
				if key.String() == "left" {
					delta = -1
				}
				n := len(field.Options)
				field.Selected = (field.Selected + delta + n) % n
				return Continue, nil
			}
		case "enter":
			if field.Kind == Completion && field.AcceptSuggestion() {
				return Continue, nil
			}
			if field.Kind == Picker {
				return Continue, f.OpenPicker()
			}
			if f.Cursor < len(f.Fields)-1 {
				return Continue, f.Move(1)
			}
			return Submitted, nil
		}
	}
	if !field.IsText() {
		return Continue, nil
	}
	var cmd tea.Cmd
	field.Input, cmd = field.Input.Update(msg)
	field.Value = field.Input.Value()
	if isKey {
		field.RefreshSuggestions()
	}
	return Continue, cmd
}

// OpenPicker opens the picker of the field under the cursor, if it is a
// picker field.
func (f *Form) OpenPicker() tea.Cmd {
	if f.Cursor < 0 || f.Cursor >= len(f.Fields) || f.Fields[f.Cursor].Kind != Picker {
		return nil
	}
	q := fields.NewTextInput()
	q.Prompt = ""
	q.Placeholder = "type to filter…"
	q.CharLimit = 0
	q.SetWidth(52)
	f.Picker = &PickerState{FieldIndex: f.Cursor, Query: q}
	f.Blur()
	f.filterPicker()
	// Start on the current choice.
	for i, index := range f.Picker.Filtered {
		if index == f.Fields[f.Cursor].Selected {
			f.Picker.Cursor = i
		}
	}
	return f.Picker.Query.Focus()
}

func (f *Form) updatePicker(msg tea.Msg) tea.Cmd {
	ps := f.Picker
	field := &f.Fields[ps.FieldIndex]
	if key, ok := msg.(tea.KeyPressMsg); ok {
		n := len(ps.Filtered)
		switch key.String() {
		case "esc":
			f.Picker = nil
			return f.focus()
		case "up", "ctrl+p":
			if n > 0 {
				ps.Cursor = (ps.Cursor - 1 + n) % n
			}
			return nil
		case "down", "ctrl+n":
			if n > 0 {
				ps.Cursor = (ps.Cursor + 1) % n
			}
			return nil
		case "enter":
			if n > 0 {
				field.Selected = ps.Filtered[ps.Cursor]
			}
			f.Picker = nil
			return f.focus()
		}
	}
	var cmd tea.Cmd
	ps.Query, cmd = ps.Query.Update(msg)
	f.filterPicker()
	return cmd
}

func (f *Form) filterPicker() {
	ps := f.Picker
	ps.Filtered = Rank(f.Fields[ps.FieldIndex].Options, ps.Query.Value())
	if ps.Cursor >= len(ps.Filtered) {
		ps.Cursor = max(0, len(ps.Filtered)-1)
	}
}
