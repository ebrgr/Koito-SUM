package handlers

import "testing"

// TestCoalescingKeySeparatesFieldBoundaries guards simultaneous distinct listens.
func TestCoalescingKeySeparatesFieldBoundaries(t *testing.T) {
	a := LbzSubmitListenPayload{TrackMeta: LbzTrackMeta{ArtistName: "A:B", TrackName: "C", ReleaseName: "D"}}
	b := LbzSubmitListenPayload{TrackMeta: LbzTrackMeta{ArtistName: "A", TrackName: "B:C", ReleaseName: "D"}}
	if buildCaolescingKey(a) == buildCaolescingKey(b) {
		t.Fatal("different metadata coalesced")
	}
	if buildCaolescingKey(a) != buildCaolescingKey(a) {
		t.Fatal("identical metadata is unstable")
	}
}
