package engine

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestListenBrainzBodyLimit covers exact and over-limit streams without mutating history.
func TestListenBrainzBodyLimit(t *testing.T) {
	a := newProfileTestApp(t)
	const max = 16 << 20
	body := `{"listen_type":"import","payload":[]}`
	for _, tc := range []struct {
		name    string
		size    int
		status  int
		unknown bool
	}{
		{"exact", max, http.StatusBadRequest, false},
		{"over", max + 1, http.StatusRequestEntityTooLarge, false},
		{"chunked over", max + 1, http.StatusRequestEntityTooLarge, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := body + strings.Repeat(" ", tc.size-len(body))
			r := httptest.NewRequest("POST", "http://koito.test/apis/listenbrainz/1/submit-listens", strings.NewReader(data))
			r.Header.Set("Authorization", "Token first-key")
			r.Header.Set("Content-Type", "application/json")
			if tc.unknown {
				r.ContentLength = -1
			}
			w := httptest.NewRecorder()
			a.router.ServeHTTP(w, r)
			require.Equal(t, tc.status, w.Code)
			var count int
			require.NoError(t, a.store.QueryRow(`SELECT COUNT(*) FROM listens`).Scan(&count))
			require.Zero(t, count)
		})
	}
}
