// Package diff computes and renders line diffs for confirmation views, such
// as a write-back preview or a sync conflict.
package diff

import (
	"strings"

	"github.com/adejoux/tuikit/textutil"
	"github.com/adejoux/tuikit/theme"
)

// Kind tells whether a line is kept, removed or added.
type Kind int

const (
	Context Kind = iota
	Remove
	Add
)

// Line is one line of a diff.
type Line struct {
	Kind Kind
	Text string
}

// Lines returns the line diff from before to after, keeping context
// unchanged lines around each change; longer unchanged runs become a single
// "…" context line. A negative context keeps every line.
func Lines(before, after string, context int) []Line {
	a := split(before)
	b := split(after)
	n, m := len(a), len(b)
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			switch {
			case a[i] == b[j]:
				dp[i][j] = dp[i+1][j+1] + 1
			case dp[i+1][j] >= dp[i][j+1]:
				dp[i][j] = dp[i+1][j]
			default:
				dp[i][j] = dp[i][j+1]
			}
		}
	}

	out := make([]Line, 0, n+m)
	for i, j := 0, 0; i < n || j < m; {
		switch {
		case i < n && j < m && a[i] == b[j]:
			out = append(out, Line{Kind: Context, Text: a[i]})
			i++
			j++
		case j >= m || (i < n && dp[i+1][j] >= dp[i][j+1]):
			out = append(out, Line{Kind: Remove, Text: a[i]})
			i++
		default:
			out = append(out, Line{Kind: Add, Text: b[j]})
			j++
		}
	}
	if context < 0 {
		return out
	}
	return trimContext(out, context)
}

// Changed reports whether lines hold at least one addition or removal.
func Changed(lines []Line) bool {
	for _, l := range lines {
		if l.Kind != Context {
			return true
		}
	}
	return false
}

// Render renders lines with "- ", "+ " and "  " prefixes in the diff styles,
// each line cut to width cells.
func Render(st theme.Styles, lines []Line, width int) string {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		text := textutil.Truncate(l.Text, max(8, width-2))
		switch l.Kind {
		case Remove:
			out = append(out, st.DiffRemove.Render("- "+text))
		case Add:
			out = append(out, st.DiffAdd.Render("+ "+text))
		default:
			out = append(out, st.DiffContext.Render("  "+text))
		}
	}
	return strings.Join(out, "\n")
}

func split(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func trimContext(lines []Line, context int) []Line {
	if len(lines) == 0 {
		return lines
	}
	keep := make([]bool, len(lines))
	for i, line := range lines {
		if line.Kind == Context {
			continue
		}
		for j := max(0, i-context); j < min(len(lines), i+context+1); j++ {
			keep[j] = true
		}
	}
	var out []Line
	gap := false
	for i, line := range lines {
		if !keep[i] {
			gap = len(out) > 0
			continue
		}
		if gap {
			out = append(out, Line{Kind: Context, Text: textutil.Ellipsis})
			gap = false
		}
		out = append(out, line)
	}
	if len(out) == 0 {
		return lines
	}
	return out
}
