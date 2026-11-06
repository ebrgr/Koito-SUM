package handlers

import (
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/listening"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
	"net/http"
)

type NowPlayingResponse = listening.NowPlaying

func NowPlayingHandler(store db.TrackStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v, err := listening.ReadNowPlaying(r.Context(), store)
		if err != nil {
			logger.FromContext(r.Context()).Err(err).Msg("now-playing read failed")
			utils.WriteError(w, "failed to fetch currently playing track", 500)
			return
		}
		utils.WriteJSON(w, 200, v)
	}
}
