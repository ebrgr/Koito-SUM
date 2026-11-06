package webui

import (
	"fmt"
	"html/template"
	"strings"
)

func loadTemplates() (*template.Template, error) {
	tmpl, err := template.New("page.html").Funcs(template.FuncMap{
		"avatarIcon": func(name string) string {
			return map[string]string{"music": "Music2", "headphones": "Headphones", "disc": "Disc3", "radio": "Radio", "user": "UserRound"}[name]
		},
		"icon": func(name string) template.HTML {
			data, err := files.ReadFile("static/icons/" + name + ".svg")
			if err != nil {
				return ""
			}
			return template.HTML(data)
		}, // Only bundled, reviewed SVGs; never catalog/user HTML.
		"tail": func(items []Media) []Media {
			if len(items) < 2 {
				return nil
			}
			return items[1:]
		},
		"first": func(items []Media) Media {
			if len(items) == 0 {
				return Media{}
			}
			return items[0]
		},
		"initials": func(s string) string {
			var letters []rune
			for _, word := range strings.Fields(s) {
				letters = append(letters, []rune(word)[0])
				if len(letters) == 2 {
					break
				}
			}
			return strings.ToUpper(string(letters))
		},
		"rank":         func(i int) string { return fmt.Sprintf("%02d", i+1) },
		"splitPeriods": func() []string { return []string{"day", "week", "month", "year", "all_time"} },
		"ago": func(i int) string {
			minutes := (i + 1) * 5
			if minutes >= 60 {
				hours := minutes / 60
				if hours == 1 {
					return "1 hour ago"
				}
				return fmt.Sprintf("%d hours ago", hours)
			}
			return fmt.Sprintf("%d minutes ago", minutes)
		},
		"history": func(items []Media) []Media {
			out := []Media{}
			if len(items) == 0 {
				return out
			}
			for i := 0; i < 28; i++ {
				out = append(out, items[i%len(items)])
			}
			return out
		},
	}).ParseFS(files, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}
