package webui

import (
	"strings"
	"testing"
)

// TestNoScriptNavigationExposesProfileAndScopedSearchLinks checks server-rendered fallbacks.
func TestNoScriptNavigationExposesProfileAndScopedSearchLinks(t *testing.T) {
	s := pageStore(t)
	h, err := NewHandler(s, Options{})
	if err != nil {
		t.Fatal(err)
	}
	body := pageRequest(h, "/alpha?period=all_time").Body.String()
	for _, want := range []string{`<noscript>`, `href="/" aria-label="Choose listening profile"`, `href="/alpha/search" aria-label="Search"`, `Appearance settings require JavaScript.`} {
		if !strings.Contains(body, want) {
			t.Fatalf("no-script navigation missing %q", want)
		}
	}
}
