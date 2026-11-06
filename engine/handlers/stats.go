package handlers

import (
	"github.com/gabehf/koito/internal/listening"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
	"net/http"
)

type StatsResponse = listening.Stats
type statsStore = listening.StatsStore

func StatsHandler(store statsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v, err := listening.ReadStats(r.Context(), store, TimeframeFromRequest(r))
		if err != nil {
			logger.FromContext(r.Context()).Err(err).Msg("statistics read failed")
			utils.WriteError(w, "failed to get statistics", 500)
			return
		}
		utils.WriteJSON(w, 200, v)
	}
}
func safeRatio(a, b int64) float32 { return listening.Ratio(a, b) }
