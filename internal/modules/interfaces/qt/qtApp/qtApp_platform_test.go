package qtapp

import "testing"

func TestSommelierPlatform(t *testing.T) {
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	for _, c := range []struct {
		env  map[string]string
		want string
	}{
		{map[string]string{"SOMMELIER_VERSION": "0.20", "DISPLAY": ":0", "WAYLAND_DISPLAY": "wayland-0"}, "xcb"},
		{map[string]string{"SOMMELIER_VERSION": "0.20", "DISPLAY": ":0", "QT_QPA_PLATFORM": "wayland"}, ""}, // the user's choice
		{map[string]string{"SOMMELIER_VERSION": "0.20", "WAYLAND_DISPLAY": "wayland-0"}, ""},                // no X server
		{map[string]string{"DISPLAY": ":0", "WAYLAND_DISPLAY": "wayland-0"}, ""},                            // not ChromeOS
	} {
		if got := sommelierPlatform(env(c.env)); got != c.want {
			t.Errorf("%v: %q, want %q", c.env, got, c.want)
		}
	}
}
