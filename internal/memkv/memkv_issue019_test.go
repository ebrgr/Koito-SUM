package memkv

import (
	"testing"
	"time"
)

// TestDeleteVersionProtectsReplacements checks stale observations including identical values.
func TestDeleteVersionProtectsReplacements(t *testing.T) {
	s := NewStore(time.Minute)
	defer s.Close()
	for _, next := range []int{1, 2} {
		s.Set("key", 1)
		_, old, ok := s.GetVersioned("key")
		if !ok {
			t.Fatal("missing original")
		}
		s.Set("key", next)
		if s.DeleteVersion("key", old) {
			t.Fatal("deleted replacement")
		}
		got, _, ok := s.GetVersioned("key")
		if !ok || got != next {
			t.Fatalf("lost replacement %v", got)
		}
	}
	_, version, ok := s.GetVersioned("key")
	if !ok || !s.DeleteVersion("key", version) {
		t.Fatal("current entry was not removed")
	}
}
