package handlers

import (
	"errors"
	"net/http"

	"github.com/gabehf/koito/engine/middleware"
	"github.com/gabehf/koito/internal/admin"
	"github.com/gabehf/koito/internal/adminweb"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

func DeleteTrackHandler(store db.TrackStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		l.Debug().Msg("DeleteTrackHandler: Received request to delete track")

		trackID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("DeleteTrackHandler: Invalid track id")
			utils.WriteError(w, "invalid track id", http.StatusBadRequest)
			return
		}

		l.Debug().Msgf("DeleteTrackHandler: Deleting track with ID %d", trackID)

		err = store.DeleteTrack(ctx, int32(trackID))
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			l.Err(err).Msg("DeleteTrackHandler: Failed to delete track")
			utils.WriteError(w, adminweb.PublicError(err), adminweb.ErrorStatus(err))
			return
		}

		l.Debug().Msgf("DeleteTrackHandler: Successfully deleted track with ID %d", trackID)
		w.WriteHeader(http.StatusNoContent)
	}
}

// DeleteTrackAliasHandler deletes an alias for a given track.
func DeleteTrackAliasHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		trackID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("DeleteTrackAliasHandler: Invalid track id")
			utils.WriteError(w, "invalid track id", http.StatusBadRequest)
			return
		}

		body, err := utils.DecodeBody[struct {
			Alias string `json:"alias"`
		}](r)
		if err != nil || body.Alias == "" {
			l.Debug().Msg("DeleteArtistAliasHandler: Invalid or missing alias in request body")
			utils.WriteError(w, "alias must be provided", http.StatusBadRequest)
			return
		}

		if err = admin.EditAlias(ctx, store, middleware.GetUserFromContext(ctx), "track", trackID, "delete", body.Alias); err != nil {
			l.Error().Err(err).Msg("DeleteTrackAliasHandler: Failed to delete track alias")
			utils.WriteError(w, adminweb.PublicError(err), adminweb.ErrorStatus(err))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// DeleteTrackArtistHandler removes an artist from a given track.
func DeleteTrackArtistHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)

		trackID, err := utils.ParseIDParam(r, "id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("RemoveTrackArtistHandler: Invalid track id")
			utils.WriteError(w, "invalid track id", http.StatusBadRequest)
			return
		}

		artistID, err := utils.ParseIDParam(r, "artist_id")
		if err != nil {
			l.Debug().AnErr("error", err).Msg("RemoveTrackArtistHandler: Invalid artist id")
			utils.WriteError(w, "invalid artist id", http.StatusBadRequest)
			return
		}

		if err = admin.ChangeTrackArtist(ctx, store, middleware.GetUserFromContext(ctx), trackID, artistID, false); err != nil {
			l.Error().Err(err).Msg("RemoveTrackArtistHandler: Failed to remove artist from track")
			utils.WriteError(w, adminweb.PublicError(err), adminweb.ErrorStatus(err))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
