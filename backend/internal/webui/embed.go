// Package webui embeds the production web build into the Go binary.
// The dist/ directory is populated by the web build step.
package webui

import (
	"embed"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

//go:embed dist
var Dist embed.FS

// StaticHandler returns an HTTP handler that serves embedded static files.
// SPA fallback: non-API requests return index.html.
func StaticHandler() http.Handler {
	sub, err := fs.Sub(Dist, "dist")
	if err != nil {
		panic("webui: failed to sub dist: " + err.Error())
	}

	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't intercept API routes
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/health/") {
			http.NotFound(w, r)
			return
		}

		// Try to serve the file directly
		path := filepath.Join(".", r.URL.Path)
		if _, err := sub.Open(path); err == nil {
			// Set content type based on extension
			contentType := http.DetectContentType([]byte("static"))
			if strings.HasSuffix(r.URL.Path, ".js") {
				contentType = "application/javascript"
			} else if strings.HasSuffix(r.URL.Path, ".css") {
				contentType = "text/css"
			} else if strings.HasSuffix(r.URL.Path, ".svg") {
				contentType = "image/svg+xml"
			} else if strings.HasSuffix(r.URL.Path, ".json") {
				contentType = "application/json"
			}
			w.Header().Set("Content-Type", contentType)
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback: serve index.html
		indexHTML, err := fs.ReadFile(sub, "dist/index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	})
}
