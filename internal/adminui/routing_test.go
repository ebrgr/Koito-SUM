package adminui

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

// TestRouteManifest checks documents and POST actions against the client contract.
func TestRouteManifest(t *testing.T) {
	data, err := os.ReadFile("../../test_assets/admin-web-ui-routes.json")
	require.NoError(t, err)
	var cases [][]any
	require.NoError(t, json.Unmarshal(data, &cases))
	for _, c := range cases {
		require.Equal(t, c[2].(bool), IsRoute(c[0].(string), c[1].(string)), c[0].(string)+" "+c[1].(string))
	}
}
