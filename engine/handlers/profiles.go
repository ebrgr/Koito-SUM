package handlers

import (
	"github.com/gabehf/koito/engine/middleware"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/utils"
	"net/http"
	"net/url"
	"strings"
)

func ListProfilesHandler(store db.ProfileStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		profiles, err := store.ListListeningProfiles(r.Context())
		if err != nil {
			utils.WriteError(w, "failed to list profiles", 500)
			return
		}
		utils.WriteJSON(w, 200, profiles)
	}
}
func GetProfileHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJSON(w, 200, middleware.ProfileFromContext(r.Context()))
	}
}
func UpdateProfileHandler(store db.ProfileStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := middleware.ProfileFromContext(r.Context())
		body, err := utils.DecodeBody[struct {
			DisplayName *string `json:"display_name"`
			Slug        *string `json:"slug"`
		}](r)
		if err != nil || (body.DisplayName == nil && body.Slug == nil) {
			utils.WriteError(w, "display_name or slug required", 400)
			return
		}
		name, slug := p.DisplayName, p.Slug
		if body.DisplayName != nil {
			name = *body.DisplayName
		}
		if body.Slug != nil {
			slug = *body.Slug
		}
		if err := store.UpdateListeningProfile(r.Context(), p.ID, name, slug); err != nil {
			utils.WriteError(w, "invalid, reserved or duplicate profile name/slug", 400)
			return
		}
		updated, err := store.GetListeningProfileByID(r.Context(), p.ID)
		if err != nil {
			utils.WriteError(w, "profile not found", 404)
			return
		}
		utils.WriteJSON(w, 200, updated)
	}
}

func SetProfileRelayHandler(store db.ProfileStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := db.ProfileID(r.Context())
		if r.Method == "DELETE" {
			if err := store.SetProfileRelay(r.Context(), id, "", ""); err != nil {
				utils.WriteError(w, "failed to remove relay", 500)
				return
			}
			w.WriteHeader(204)
			return
		}
		body, err := utils.DecodeBody[struct {
			URL   string `json:"url"`
			Token string `json:"token"`
		}](r)
		parsed, parseErr := url.Parse(body.URL)
		if err != nil || parseErr != nil || parsed.Host == "" || parsed.User != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.RawQuery != "" || parsed.Fragment != "" || body.Token == "" {
			utils.WriteError(w, "valid relay URL and token required", 400)
			return
		}
		if err := store.SetProfileRelay(r.Context(), id, strings.TrimRight(body.URL, "/"), body.Token); err != nil {
			utils.WriteError(w, "failed to configure relay", 500)
			return
		}
		w.WriteHeader(204)
	}
}
