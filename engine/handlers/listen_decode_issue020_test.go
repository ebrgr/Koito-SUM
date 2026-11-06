package handlers

import (
	"errors"
	"strings"
	"testing"
)

// TestListenDecoderStopsAtItemLimit rejects the 1001st item before parsing it.
func TestListenDecoderStopsAtItemLimit(t *testing.T) {
	raw := `{"listen_type":"import","payload":[` + strings.Repeat(`{},`, 1000) + `INVALID]}`
	_, err := decodeListenRequest(strings.NewReader(raw))
	if !errors.Is(err, errTooManyListens) {
		t.Fatalf("got %v", err)
	}
}

// TestListenDecoderCompatibilityPreservesSupportedEnvelopes covers case and repeated fields.
func TestListenDecoderCompatibilityPreservesSupportedEnvelopes(t *testing.T) {
	for _, tc := range []struct {
		body string
		want int
		kind LbzListenType
		bad  bool
	}{
		{`{"LISTEN_TYPE":"import","Payload":[{},{}]}`, 2, ListenTypeImport, false},
		{`{"payload":[{}],"payload":[{},{}]}`, 2, "", false},
		{`{"payload":[{}],"payload":null}`, 0, "", false},
		{`{"payload":{},"listen_type":"single"}`, 0, "", true},
		{`{"payload":[` + strings.Repeat(`{},`, 999) + `{}]}`, 1000, "", false},
	} {
		got, err := decodeListenRequest(strings.NewReader(tc.body))
		if (err != nil) != tc.bad || len(got.Payload) != tc.want || got.ListenType != tc.kind {
			t.Fatalf("unexpected result: size=%d kind=%s err=%v", len(got.Payload), got.ListenType, err)
		}
	}
}
