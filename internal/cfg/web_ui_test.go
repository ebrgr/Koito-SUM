package cfg

import "testing"

// TestWebUIConfig checks the default, supported values and invalid modes.
func TestWebUIConfig(t *testing.T) {
	for _, tc := range []struct{ value, want string }{{"", "go"}, {"go", "go"}, {"GO", "go"}} {
		t.Run(tc.value, func(t *testing.T) {
			c, err := loadConfig(func(key string) string {
				if key == WEB_UI_ENV {
					return tc.value
				}
				return ""
			}, "test")
			if err != nil || c.webUI != tc.want {
				t.Fatalf("mode %q: got %q, error %v", tc.value, c.webUI, err)
			}
		})
	}
	for _, invalid := range []string{"typo", "react", "REACT"} {
		if _, err := loadConfig(func(key string) string {
			if key == WEB_UI_ENV {
				return invalid
			}
			return ""
		}, "test"); err == nil {
			t.Fatal("invalid mode accepted")
		}
	}
}
