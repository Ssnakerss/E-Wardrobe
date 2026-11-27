// Package webui embeds the production web build into the Go binary.
// The dist/ directory is populated by copying from web/dist before go build.
package webui

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed dist
var Dist embed.FS

// subFS is the embedded filesystem without the "dist/" prefix.
var subFS, _ = fs.Sub(Dist, "dist")

// StaticHandler serves embedded static files with proper caching and SPA fallback.
func StaticHandler() http.Handler {
	fileServer := http.FileServer(http.FS(subFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't intercept API or health routes
		if strings.HasPrefix(r.URL.Path, "/api/") ||
			strings.HasPrefix(r.URL.Path, "/health/") ||
			strings.HasPrefix(r.URL.Path, "/sw.") {
			http.NotFound(w, r)
			return
		}

		// Normalize: redirect trailing-slash directories
		requestPath := r.URL.Path
		if requestPath != "/" && strings.HasSuffix(requestPath, "/") {
			http.Redirect(w, r, strings.TrimRight(requestPath, "/"), http.StatusTemporaryRedirect)
			return
		}

		// Try to find the file in embed
		filePath := strings.TrimPrefix(requestPath, "/")
		if filePath == "" {
			filePath = "index.html"
		}

		f, err := subFS.Open(filePath)
		if err == nil {
			f.Close()

			// Set cache headers based on file type
			if filePath == "index.html" || strings.HasPrefix(filePath, "index.html/") {
				setNoCacheHeaders(w)
			} else if isHashedAsset(requestPath) {
				setImmutableCacheHeaders(w)
			} else {
				setShortCacheHeaders(w)
			}

			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback: return index.html
		idx, err := fs.ReadFile(subFS, "index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		setNoCacheHeaders(w)
		w.Write(idx)
	})
}

// isHashedAsset returns true if the path looks like a Vite-built hashed asset.
func isHashedAsset(p string) bool {
	base := path.Base(p)
	idx := strings.LastIndex(base, ".")
	if idx <= 0 {
		return false
	}
	name := base[:idx]
	if dash := strings.LastIndex(name, "-"); dash > 0 {
		suffix := name[dash+1:]
		if len(suffix) >= 8 {
			return true
		}
	}
	return false
}

func setNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
}

func setImmutableCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
}

func setShortCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "public, max-age=3600")
}
