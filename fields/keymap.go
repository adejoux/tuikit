// Package fields holds the text field keymaps that follow tuikit's focus
// rules: readline editing keys stay with the field, while ctrl+n and ctrl+p
// are left free for the command palette and pickers.
package fields

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
)

// Reserved lists the keys that belong to text fields. Applications never
// bind an action to them, so editing a value always works as in a shell.
var Reserved = []string{
	"ctrl+a", "ctrl+b", "ctrl+d", "ctrl+e", "ctrl+f", "ctrl+h",
	"ctrl+k", "ctrl+t", "ctrl+u", "ctrl+v", "ctrl+w",
}

// IsReserved reports whether k, a key string such as "ctrl+w" or
// "alt+backspace", is kept for text editing.
func IsReserved(k string) bool {
	if len(k) > 4 && k[:4] == "alt+" {
		return true
	}
	for _, r := range Reserved {
		if r == k {
			return true
		}
	}
	return false
}

// TextInputKeyMap is bubbles' textinput keymap without ctrl+n and ctrl+p:
// suggestions move with the arrows only.
func TextInputKeyMap() textinput.KeyMap {
	km := textinput.DefaultKeyMap()
	km.NextSuggestion = key.NewBinding(key.WithKeys("down"))
	km.PrevSuggestion = key.NewBinding(key.WithKeys("up"))
	return km
}

// TextAreaKeyMap is bubbles' textarea keymap without ctrl+n and ctrl+p: lines
// move with the arrows only.
func TextAreaKeyMap() textarea.KeyMap {
	km := textarea.DefaultKeyMap()
	km.LineNext = key.NewBinding(key.WithKeys("down"), key.WithHelp("down", "next line"))
	km.LinePrevious = key.NewBinding(key.WithKeys("up"), key.WithHelp("up", "previous line"))
	return km
}

// NewTextInput returns a textinput using TextInputKeyMap.
func NewTextInput() textinput.Model {
	in := textinput.New()
	in.KeyMap = TextInputKeyMap()
	return in
}

// NewTextArea returns a textarea using TextAreaKeyMap.
func NewTextArea() textarea.Model {
	ta := textarea.New()
	ta.KeyMap = TextAreaKeyMap()
	return ta
}
