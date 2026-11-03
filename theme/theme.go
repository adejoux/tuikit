// Package theme holds the colour roles and the lipgloss styles shared by the
// tools built on tuikit. Applications add their own domain colours on top of
// these roles; the roles here are the ones every screen needs.
package theme

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// Theme is a named set of colour roles.
type Theme struct {
	Name       string
	Background color.Color
	Surface    color.Color
	SurfaceAlt color.Color
	Border     color.Color
	Text       color.Color
	Muted      color.Color
	Subtle     color.Color // between Text and Muted, used for key names
	Accent     color.Color
	Secondary  color.Color // second accent, end of the dialog title gradient
	Success    color.Color
	Warning    color.Color
	Danger     color.Color
}

// Dark is the default theme.
func Dark() Theme {
	return Theme{
		Name:       "dark",
		Background: lipgloss.Color("#0B1020"),
		Surface:    lipgloss.Color("#121A2B"),
		SurfaceAlt: lipgloss.Color("#162035"),
		Border:     lipgloss.Color("#334155"),
		Text:       lipgloss.Color("#E5E7EB"),
		Muted:      lipgloss.Color("#8290A6"),
		Subtle:     lipgloss.Color("#AEB8C8"),
		Accent:     lipgloss.Color("#38BDF8"),
		Secondary:  lipgloss.Color("#C084FC"),
		Success:    lipgloss.Color("#4ADE80"),
		Warning:    lipgloss.Color("#FBBF24"),
		Danger:     lipgloss.Color("#FB7185"),
	}
}

// Light is the light theme.
func Light() Theme {
	return Theme{
		Name:       "light",
		Background: lipgloss.Color("#F7F8FA"),
		Surface:    lipgloss.Color("#FFFFFF"),
		SurfaceAlt: lipgloss.Color("#EEF2F7"),
		Border:     lipgloss.Color("#CBD5E1"),
		Text:       lipgloss.Color("#18212F"),
		Muted:      lipgloss.Color("#64748B"),
		Subtle:     lipgloss.Color("#475569"),
		Accent:     lipgloss.Color("#2563EB"),
		Secondary:  lipgloss.Color("#7C3AED"),
		Success:    lipgloss.Color("#15803D"),
		Warning:    lipgloss.Color("#B45309"),
		Danger:     lipgloss.Color("#DC2626"),
	}
}

// ByName returns the theme with the given name, or Dark for any other name.
func ByName(name string) Theme {
	if name == "light" {
		return Light()
	}
	return Dark()
}

// Styles are the lipgloss styles derived from a Theme.
type Styles struct {
	Panel           lipgloss.Style
	ActivePanel     lipgloss.Style
	Tab             lipgloss.Style
	ActiveTab       lipgloss.Style
	SelectionMarker lipgloss.Style
	ListHeader      lipgloss.Style
	Muted           lipgloss.Style
	Title           lipgloss.Style
	Heading         lipgloss.Style
	Search          lipgloss.Style
	Status          lipgloss.Style
	Error           lipgloss.Style
	Success         lipgloss.Style
	Warning         lipgloss.Style
	Chip            lipgloss.Style
	HeaderBrand     lipgloss.Style
	HeaderMetric    lipgloss.Style

	// Key hints, after Crush: keys in a subtle tone, descriptions one step
	// dimmer and a discreet separator, so hints never compete with content.
	HelpKey       lipgloss.Style
	HelpDesc      lipgloss.Style
	HelpAction    lipgloss.Style // action names in the shortcut dialog
	HelpSeparator lipgloss.Style

	// Dialogs: rounded accent frame and a title followed by a ╱╱╱ gradient
	// from DialogGradFrom to DialogGradTo.
	Dialog         lipgloss.Style
	DialogTitle    lipgloss.Style
	DialogGradFrom color.Color
	DialogGradTo   color.Color
	DialogPill     lipgloss.Style
	DialogPillOn   lipgloss.Style
	DialogPrompt   lipgloss.Style

	DiffBox     lipgloss.Style
	DiffAdd     lipgloss.Style
	DiffRemove  lipgloss.Style
	DiffContext lipgloss.Style
}

// NewStyles derives the shared styles from t.
func NewStyles(t Theme) Styles {
	return Styles{
		Panel:           lipgloss.NewStyle().Foreground(t.Text).Border(lipgloss.RoundedBorder()).BorderForeground(t.Border).Padding(0, 1),
		ActivePanel:     lipgloss.NewStyle().Foreground(t.Text).Border(lipgloss.RoundedBorder()).BorderForeground(t.Accent).Padding(0, 1),
		Tab:             lipgloss.NewStyle().Foreground(t.Muted).Padding(0, 1),
		ActiveTab:       lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Padding(0, 1),
		SelectionMarker: lipgloss.NewStyle().Foreground(t.Accent).Bold(true),
		ListHeader:      lipgloss.NewStyle().Foreground(t.Muted).Background(t.SurfaceAlt),
		Muted:           lipgloss.NewStyle().Foreground(t.Muted),
		Title:           lipgloss.NewStyle().Foreground(t.Text).Bold(true),
		Heading:         lipgloss.NewStyle().Foreground(t.Accent).Bold(true),
		Search:          lipgloss.NewStyle().Foreground(t.Accent).Bold(true),
		Status:          lipgloss.NewStyle().Foreground(t.Muted),
		Error:           lipgloss.NewStyle().Foreground(t.Danger).Bold(true),
		Success:         lipgloss.NewStyle().Foreground(t.Success).Bold(true),
		Warning:         lipgloss.NewStyle().Foreground(t.Warning).Bold(true),
		Chip:            lipgloss.NewStyle().Foreground(t.Text).Background(t.SurfaceAlt).Padding(0, 1),
		HeaderBrand:     lipgloss.NewStyle().Foreground(t.Accent).Bold(true),
		HeaderMetric:    lipgloss.NewStyle().Foreground(t.Muted),
		HelpKey:         lipgloss.NewStyle().Foreground(t.Subtle),
		HelpDesc:        lipgloss.NewStyle().Foreground(t.Muted),
		HelpAction:      lipgloss.NewStyle().Foreground(t.Text),
		HelpSeparator:   lipgloss.NewStyle().Foreground(t.Border),
		Dialog:          lipgloss.NewStyle().Foreground(t.Text).Border(lipgloss.RoundedBorder()).BorderForeground(t.Accent).Padding(0, 1),
		DialogTitle:     lipgloss.NewStyle().Foreground(t.Accent).Bold(true),
		DialogGradFrom:  t.Accent,
		DialogGradTo:    t.Secondary,
		DialogPill:      lipgloss.NewStyle().Foreground(t.Muted).Padding(0, 1),
		DialogPillOn:    lipgloss.NewStyle().Foreground(t.Surface).Background(t.Accent).Bold(true).Padding(0, 1),
		DialogPrompt:    lipgloss.NewStyle().Foreground(t.Accent).Bold(true),
		DiffBox:         lipgloss.NewStyle().Foreground(t.Text).Border(lipgloss.RoundedBorder()).BorderForeground(t.Border).Padding(0, 1),
		DiffAdd:         lipgloss.NewStyle().Foreground(t.Success),
		DiffRemove:      lipgloss.NewStyle().Foreground(t.Danger),
		DiffContext:     lipgloss.NewStyle().Foreground(t.Muted),
	}
}

// Gutter is the two-cell selection gutter in front of a row: an accent bar
// when selected, blank otherwise.
func (s Styles) Gutter(selected bool) string {
	if selected {
		return s.SelectionMarker.Render("▌") + " "
	}
	return "  "
}

// SelectedLine renders value as the selected row of a list of the given
// width: gutter, then the value cut or padded to the remaining width.
func (s Styles) SelectedLine(value string, width int) string {
	content := max(0, width-2)
	if lipgloss.Width(value) > content {
		value = ansi.Truncate(value, content, "…")
	}
	return s.Gutter(true) + value + strings.Repeat(" ", max(0, content-lipgloss.Width(value)))
}

// Line renders value as an unselected row of the given width.
func (s Styles) Line(value string, width int) string {
	content := max(0, width-2)
	if lipgloss.Width(value) > content {
		value = ansi.Truncate(value, content, "…")
	}
	return s.Gutter(false) + value
}
