package server

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const readinessTimeout = 2 * time.Second

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeStatus(w, http.StatusOK, "ok")
}

func readyzHandler(logger *slog.Logger, checks []ReadinessCheck) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), readinessTimeout)
		defer cancel()
		for _, check := range checks {
			if err := check(ctx); err != nil {
				logger.Warn("readiness check failed", "err", err)
				writeStatus(w, http.StatusServiceUnavailable, "not ready")
				return
			}
		}
		writeStatus(w, http.StatusOK, "ready")
	}
}

func writeStatus(w http.ResponseWriter, code int, body string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_, _ = io.WriteString(w, body+"\n")
}
