package imagefetch

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestDownloadBoundsAndCancellation exercises slow and oversized local sources.
func TestDownloadBoundsAndCancellation(t *testing.T) {
	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/slow":
			close(started)
			<-r.Context().Done()
		case "/large":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(append([]byte("\x89PNG\r\n\x1a\n"), []byte(strings.Repeat("x", MaxBytes))...))
		case "/redirect":
			http.Redirect(w, r, "/redirect", http.StatusFound)
		}
	}))
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { _, err := Download(ctx, server.URL+"/slow"); done <- err }()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		cancel()
		t.Fatal("image request did not start")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancellation blocked")
	}
	if _, err := Download(context.Background(), server.URL+"/large"); err == nil {
		t.Fatal("accepted oversized image")
	}
	if _, err := Download(context.Background(), server.URL+"/redirect"); err == nil {
		t.Fatal("accepted redirect loop")
	}
}
