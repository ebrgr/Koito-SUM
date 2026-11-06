package webui

import (
	"fmt"
	"io/fs"
	"net/http"
	"sort"
)

// PublicFiles keeps legacy public URLs available without a frontend directory.
func PublicFiles() http.FileSystem {
	sub, err := fs.Sub(files, "static/public")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}

// Assets exposes only presentation assets, never demo data or fake API routes.
func Assets() http.Handler {
	mux := http.NewServeMux()
	sub, _ := fs.Sub(files, "static")
	mux.HandleFunc("GET /themes.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data, _ := files.ReadFile("palettes.json")
		_, _ = w.Write(data)
	})
	mux.HandleFunc("GET /themes.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		for _, name := range themeNames {
			fmt.Fprintf(w, "[data-theme=%q]{", name)
			keys := []string{}
			for k := range palettes[name] {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Fprintf(w, "--color-%s:%s;", cssKey(k), palettes[name][k])
			}
			fmt.Fprintln(w, "}")
		}
	})
	mux.Handle("GET /", http.FileServer(http.FS(sub)))
	return mux
}
