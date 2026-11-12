package server

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/webb1es/foosballcup/internal/tournament"
)

// SnapshotLoader returns the current public view of the tournament.
type SnapshotLoader func(ctx context.Context) (tournament.Snapshot, error)

func tournamentHandler(logger *slog.Logger, load SnapshotLoader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap, err := load(r.Context())
		if errors.Is(err, tournament.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, errorBody{Error: "no tournament has been set up"})
			return
		}
		if err != nil {
			logger.Error("load tournament", "err", err)
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: "internal error"})
			return
		}
		writeJSON(w, http.StatusOK, snap)
	}
}

type errorBody struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
