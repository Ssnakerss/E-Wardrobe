// Package webui embeds the production web build into the Go binary.
// The dist/ directory is populated by `make web-build` or by copying
// from the web/ directory before `go build`.
package webui

import (
	"embed"
	"net/http"
	"path"
	"strings"
)

//go:embed dist
var Dist embed.FS

// StaticHandler returns an HTTP handler that serves embedded static files
// with proper caching, content types, and SPA fallback.
func StaticHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't intercept API or health routes
		if strings.HasPrefix(r.URL.Path, "/api/") ||
			strings.HasPrefix(r.URL.Path, "/health/") ||
			strings.HasPrefix(r.URL.Path, "/sw.") {
			http.NotFound(w, r)
			return
		}

		// Normalise path: strip trailing slash for directory requests
		requestPath := r.URL.Path
		if requestPath != "/" && strings.HasSuffix(requestPath, "/") {
			requestPath = strings.TrimRight(requestPath, "/")
		}

		// Try to serve the requested file
		f, err := Dist.Open(strings.TrimPrefix(requestPath, "/"))
		if err == nil {
			f.Close()
			// index.html → no-cache
			if requestPath == "/" || strings.HasSuffix(requestPath, "/index.html") {
				setNoCacheHeaders(w)
				http.ServeFile(w, r, strings.TrimPrefix(requestPath, "/"))
				return
			}

			// Hashed assets (JS/CSS) → immutable
			if isHashedAsset(requestPath) {
				setImmutableCacheHeaders(w)
				http.ServeFile(w, r, strings.TrimPrefix(requestPath, "/"))
				return
			}

			// Other static assets (images, favicon, robots) → 1h cache
			setShortCacheHeaders(w)
			http.ServeFile(w, r, strings.TrimPrefix(requestPath, "/"))
			return
		}

		// SPA fallback: serve index.html for any non-file route
		idx, err := Dist.ReadFile("index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// Detect whether the original request was a directory
		// (e.g. /items/ → should redirect to /items)
		if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
			// Redirect to strip trailing slash
			target := strings.TrimRight(r.URL.Path, "/")
			http.Redirect(w, r, target, http.StatusTemporaryRedirect)
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
	// Vite hashes: name-contenthash.ext (contains a dash between name and hash)
	idx := strings.LastIndex(base, ".")
	if idx <= 0 {
		return false
	}
	name := base[:idx]
	// Check for content hash pattern: at least 8 hex chars after last dash
	if dash := strings.LastIndex(name, "-"); dash > 0 {
		suffix := name[dash+1:]
		if len(suffix) >= 8 {
			return true
		}
	}
	return false
}

// setNoCacheHeaders sets Cache-Control: no-cache, no-store, must-revalidate.
func setNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
}

// setImmutableCacheHeaders sets Cache-Control: public, max-age=31536000, immutable.
func setImmutableCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
}

// setShortCacheHeaders sets Cache-Control: public, max-age=3600.
func setShortCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "public, max-age=3600")
}
