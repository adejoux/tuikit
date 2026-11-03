package theme

import (
	"reflect"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestThemesSetEveryRole(t *testing.T) {
	for _, th := range []Theme{Dark(), Light()} {
		v := reflect.ValueOf(th)
		for i := 0; i < v.NumField(); i++ {
			if v.Field(i).IsZero() {
				t.Errorf("%s theme: %s is not set", th.Name, v.Type().Field(i).Name)
			}
		}
	}
}

func TestByNameFallsBackToDark(t *testing.T) {
	if ByName("light").Name != "light" || ByName("dark").Name != "dark" || ByName("solarized").Name != "dark" {
		t.Fatal("ByName must return light, else dark")
	}
}

func TestStylesUseThemeColours(t *testing.T) {
	th := Dark()
	st := NewStyles(th)
	if st.DialogGradFrom != th.Accent || st.DialogGradTo != th.Secondary {
		t.Fatal("dialog gradient must run from Accent to Secondary")
	}
	if st.ActivePanel.GetBorderTopForeground() != th.Accent {
		t.Fatal("the focused panel border must use Accent")
	}
}

func TestSelectedLineFillsWidth(t *testing.T) {
	st := NewStyles(Dark())
	for _, width := range []int{2, 10, 40} {
		for _, value := range []string{"", "short", "a value much longer than the row width"} {
			if got := lipgloss.Width(st.SelectedLine(value, width)); got != width {
				t.Fatalf("SelectedLine(%q, %d) is %d cells", value, width, got)
			}
			if got := lipgloss.Width(st.Line(value, width)); got > width {
				t.Fatalf("Line(%q, %d) is %d cells", value, width, got)
			}
		}
	}
}
