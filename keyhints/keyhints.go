// Package keyhints renders one-line key hints in the style of Crush's short
// help: "key desc • key desc • …".
package keyhints

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/adejoux/tuikit/theme"
)

// Hint is one "key description" pair.
type Hint struct {
	Key, Desc string
}

const (
	// Separator sits between two hints.
	Separator = " • "
	ellipsis  = "…"
)

// More is the usual pinned hint: it keeps the way to discover every shortcut
// visible however narrow the terminal is.
var More = Hint{Key: "?", Desc: "more"}

// Render packs hints on a single line of at most width cells. Hints that do
// not fit are replaced by an ellipsis, and the line never ends on a
// separator. pinned, when set, is always kept at the end.
func Render(st theme.Styles, hints []Hint, pinned *Hint, width int) string {
	if width <= 0 {
		return ""
	}
	sep := st.HelpSeparator.Render(Separator)
	segment := func(h Hint) string {
		return st.HelpKey.Render(h.Key) + " " + st.HelpDesc.Render(h.Desc)
	}

	tail := ""
	if pinned != nil {
		tail = segment(*pinned)
		if lipgloss.Width(tail) > width {
			return ansi.Truncate(tail, width, ellipsis)
		}
	}
	budget := width
	if tail != "" {
		budget -= lipgloss.Width(sep) + lipgloss.Width(tail)
	}

	cut := st.HelpSeparator.Render(ellipsis)
	var b strings.Builder
	used := 0
	for i, h := range hints {
		seg := segment(h)
		if used > 0 {
			seg = sep + seg
		}
		w := lipgloss.Width(seg)
		// Keep room for " …" unless this is the last hint, so a cut line
		// always says that more hints exist.
		reserve := 0
		if i < len(hints)-1 {
			reserve = 1 + lipgloss.Width(cut)
		}
		if used+w+reserve > budget {
			if used > 0 {
				b.WriteString(" ")
				used++
			}
			if used+lipgloss.Width(cut) <= budget {
				b.WriteString(cut)
				used += lipgloss.Width(cut)
			}
			break
		}
		b.WriteString(seg)
		used += w
	}
	if tail != "" {
		if used > 0 {
			b.WriteString(sep)
		}
		b.WriteString(tail)
	}
	return b.String()
}

// Pairs renders alternating key and description strings as a hint line.
// width <= 0 means no width limit.
func Pairs(st theme.Styles, width int, pairs ...string) string {
	list := make([]Hint, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		list = append(list, Hint{Key: pairs[i], Desc: pairs[i+1]})
	}
	if width <= 0 {
		width = 1 << 16
	}
	return Render(st, list, nil, width)
}
