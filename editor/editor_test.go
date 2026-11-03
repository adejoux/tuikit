package editor

import (
	"errors"
	"reflect"
	"testing"
)

func TestArgsPositionKnownEditors(t *testing.T) {
	cases := []struct {
		editor string
		line   int
		want   []string
	}{
		{"nvim", 12, []string{"+12", "f.yaml"}},
		{"/usr/bin/vim -u NONE", 3, []string{"+3", "f.yaml"}},
		{"hx", 7, []string{"+7", "f.yaml"}},
		{"code --wait", 9, []string{"-g", "f.yaml:9"}},
		{"subl -w", 2, []string{"f.yaml:2"}},
		{"ed", 4, []string{"f.yaml"}},
		{"nvim", 0, []string{"f.yaml"}},
		{"", 5, []string{"f.yaml"}},
	}
	for _, c := range cases {
		if got := Args(c.editor, "f.yaml", c.line); !reflect.DeepEqual(got, c.want) {
			t.Errorf("Args(%q, %d) = %v, want %v", c.editor, c.line, got, c.want)
		}
	}
}

func TestCommandPrefersVisual(t *testing.T) {
	t.Setenv("VISUAL", "nvim")
	t.Setenv("EDITOR", "nano")
	if got := Command(); got != "nvim" {
		t.Fatalf("Command = %q, want nvim", got)
	}
	t.Setenv("VISUAL", " ")
	if got := Command(); got != "nano" {
		t.Fatalf("Command = %q, want nano", got)
	}
}

func TestOpenWithoutEditorReportsError(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")
	msg := Open("f.yaml", 1)()
	done, ok := msg.(FinishedMsg)
	if !ok || !errors.Is(done.Err, ErrNoEditor) || done.Path != "f.yaml" {
		t.Fatalf("msg = %#v", msg)
	}
}
