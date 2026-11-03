package fields

import (
	"reflect"
	"testing"

	"charm.land/bubbles/v2/key"
)

// bindings returns every key of every binding field of a keymap struct.
func bindings(km any) map[string]bool {
	keys := map[string]bool{}
	v := reflect.ValueOf(km)
	for i := 0; i < v.NumField(); i++ {
		if b, ok := v.Field(i).Interface().(key.Binding); ok {
			for _, k := range b.Keys() {
				keys[k] = true
			}
		}
	}
	return keys
}

func TestKeymapsLeaveCtrlNAndCtrlPFree(t *testing.T) {
	for name, km := range map[string]any{"textinput": TextInputKeyMap(), "textarea": TextAreaKeyMap()} {
		keys := bindings(km)
		for _, k := range []string{"ctrl+n", "ctrl+p"} {
			if keys[k] {
				t.Errorf("%s keymap still binds %s", name, k)
			}
		}
		for _, k := range []string{"up", "down", "ctrl+w", "ctrl+u"} {
			if !keys[k] {
				t.Errorf("%s keymap lost %s", name, k)
			}
		}
	}
}

func TestApplicationKeysAreNotUsedByFields(t *testing.T) {
	// Keys that tuikit applications act on while a field has focus.
	app := []string{"ctrl+s", "ctrl+r", "ctrl+o", "ctrl+l", "ctrl+p", "ctrl+x", "ctrl+space", "f1"}
	for name, km := range map[string]any{"textinput": TextInputKeyMap(), "textarea": TextAreaKeyMap()} {
		keys := bindings(km)
		for _, k := range app {
			if keys[k] {
				t.Errorf("%s keymap binds %s, which applications need", name, k)
			}
			if IsReserved(k) {
				t.Errorf("%s is both an application key and reserved", k)
			}
		}
	}
}

func TestReservedCoversFieldEditingKeys(t *testing.T) {
	for _, k := range []string{"ctrl+w", "ctrl+u", "ctrl+k", "alt+backspace", "alt+d"} {
		if !IsReserved(k) {
			t.Errorf("%s should be reserved", k)
		}
	}
	if IsReserved("ctrl+s") {
		t.Error("ctrl+s is an application key")
	}
	in := NewTextInput()
	if bindings(in.KeyMap)["ctrl+n"] {
		t.Error("NewTextInput must use the tuikit keymap")
	}
	ta := NewTextArea()
	if bindings(ta.KeyMap)["ctrl+p"] {
		t.Error("NewTextArea must use the tuikit keymap")
	}
}
