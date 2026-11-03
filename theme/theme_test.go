package theme

import (
	"reflect"
	"testing"
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
