package listening

import (
	"context"
	"errors"
	"fmt"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/memkv"
	"github.com/gabehf/koito/internal/models"
)

type NowPlaying struct {
	CurrentlyPlaying bool         `json:"currently_playing"`
	Track            models.Track `json:"track"`
}

// ReadNowPlaying clears only the stale cache entry observed before the lookup.
func ReadNowPlaying(ctx context.Context, s db.TrackStore) (NowPlaying, error) {
	key := fmt.Sprintf("now-playing:%d", db.ProfileID(ctx))
	value, version, ok := memkv.Store.GetVersioned(key)
	if !ok {
		return NowPlaying{}, nil
	}
	id, ok := value.(int32)
	if !ok {
		return NowPlaying{}, fmt.Errorf("invalid now-playing cache value")
	}
	track, err := s.GetTrack(ctx, db.GetTrackOpts{ID: id, AllowUnplayed: true})
	if errors.Is(err, db.ErrNotFound) {
		memkv.Store.DeleteVersion(key, version)
		return NowPlaying{}, nil
	}
	if err != nil {
		return NowPlaying{}, err
	}
	return NowPlaying{true, *track}, nil
}
