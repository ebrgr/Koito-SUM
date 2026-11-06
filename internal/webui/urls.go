package webui

import (
	"fmt"
	"net/url"
	"strings"
)

func mediaURL(slug, kind string, id int32) string {
	return fmt.Sprintf("/%s/%s/%d", url.PathEscape(slug), kind, id)
}
func queryURL(path string, q url.Values) string {
	s := q.Encode()
	if s != "" {
		return path + "?" + s
	}
	return path
}
func cloneQuery(q url.Values) url.Values {
	v := url.Values{}
	for key, values := range q {
		v[key] = append([]string(nil), values...)
	}
	return v
}
func switchURL(slug, path string, q url.Values) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	suffix := ""
	if len(parts) > 1 {
		candidate := strings.Join(parts[1:], "/")
		switch candidate {
		case "listens", "rewind", "chart/top/artists", "chart/top/albums", "chart/top/tracks":
			suffix = "/" + candidate
		}
	}
	next := cloneQuery(q)
	if suffix == "" {
		for _, key := range []string{"year", "month", "week", "from", "to"} {
			next.Del(key)
		}
	}
	for _, key := range []string{"artist_id", "album_id", "track_id", "page"} {
		next.Del(key)
	}
	return queryURL("/"+url.PathEscape(slug)+suffix, next)
}
