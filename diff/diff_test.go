package diff

import (
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/adejoux/tuikit/theme"
)

func TestLinesFindsChangesWithContext(t *testing.T) {
	before := "a\nb\nc\nd\ne\nf\ng\nh\ni\nj"
	after := "a\nb\nc\nd\ne\nF\ng\nh\ni\nj"
	got := Lines(before, after, 1)
	// Unchanged lines before the first hunk and after the last one are
	// dropped: an ellipsis only separates hunks.
	want := []Line{
		{Context, "e"},
		{Remove, "f"},
		{Add, "F"},
		{Context, "g"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Lines = %#v\nwant %#v", got, want)
	}
	if !Changed(got) {
		t.Fatal("Changed must report the edit")
	}
}

func TestLinesSeparatesDistantHunks(t *testing.T) {
	before := "1\n2\n3\n4\n5\n6\n7\n8"
	after := "X\n2\n3\n4\n5\n6\n7\nY"
	got := Lines(before, after, 1)
	var texts []string
	for _, l := range got {
		texts = append(texts, l.Text)
	}
	if strings.Join(texts, ",") != "1,X,2,…,7,8,Y" {
		t.Fatalf("hunks = %v", texts)
	}
}

func TestLinesIdenticalAndFullContext(t *testing.T) {
	if Changed(Lines("same\ntext", "same\r\ntext", 3)) {
		t.Fatal("CRLF differences must not count as changes")
	}
	all := Lines("a\nb\nc", "a\nB\nc", -1)
	if len(all) != 4 {
		t.Fatalf("negative context keeps every line, got %#v", all)
	}
}

func TestRenderPrefixes(t *testing.T) {
	st := theme.NewStyles(theme.Dark())
	out := ansi.Strip(Render(st, []Line{{Context, "a"}, {Remove, "b"}, {Add, "c"}}, 40))
	if out != "  a\n- b\n+ c" {
		t.Fatalf("render = %q", out)
	}
}
