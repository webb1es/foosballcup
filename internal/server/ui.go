package server

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// Security: the UI only loads its own scripts, styles and fonts. Avatars may be remote images.
const contentSecurityPolicy = "default-src 'self'; img-src 'self' data: https:; style-src 'self'; font-src 'self'; " +
	"connect-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'"

// uiHandler serves the built React app. Unknown paths without a file extension get
// index.html so client-side routes like /fixtures load directly.
func uiHandler(ui fs.FS) http.Handler {
	files := http.FileServerFS(ui)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")

		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name != "" && name != "index.html" {
			info, err := fs.Stat(ui, name)
			switch {
			case err == nil && !info.IsDir():
				if strings.HasPrefix(name, "assets/") {
					// Vite puts a content hash in every file name under assets/, so they never change.
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				} else {
					w.Header().Set("Cache-Control", "no-cache")
				}
				files.ServeHTTP(w, r)
				return
			case path.Ext(name) != "":
				http.NotFound(w, r)
				return
			}
		}
		serveIndex(w, ui)
	})
}

func serveIndex(w http.ResponseWriter, ui fs.FS) {
	page, err := fs.ReadFile(ui, "index.html")
	if err != nil {
		writeStatus(w, http.StatusServiceUnavailable, "The web UI has not been built. Run npm run build in web/.")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Security-Policy", contentSecurityPolicy)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(page)
}

func apiNotFound(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusNotFound, errorBody{Error: "not found"})
}
