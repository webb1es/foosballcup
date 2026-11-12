package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/webb1es/foosballcup/internal/tournament"
)

func TestTournamentHandler(t *testing.T) {
	snap := tournament.Snapshot{Tournament: tournament.Info{Name: "Entelect Zim Foosball Cup"}}

	tests := []struct {
		name     string
		load     SnapshotLoader
		wantCode int
		wantBody string
	}{
		{
			name:     "returns the snapshot",
			load:     func(context.Context) (tournament.Snapshot, error) { return snap, nil },
			wantCode: http.StatusOK,
			wantBody: `"name":"Entelect Zim Foosball Cup"`,
		},
		{
			name: "no tournament yet",
			load: func(context.Context) (tournament.Snapshot, error) {
				return tournament.Snapshot{}, tournament.ErrNotFound
			},
			wantCode: http.StatusNotFound,
			wantBody: `"error":"no tournament has been set up"`,
		},
		{
			name: "database failure hides the cause",
			load: func(context.Context) (tournament.Snapshot, error) {
				return tournament.Snapshot{}, errors.New("disk I/O error")
			},
			wantCode: http.StatusInternalServerError,
			wantBody: `"error":"internal error"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			routes(discardLogger(), Deps{LoadSnapshot: tt.load}).
				ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/tournament", nil))

			if rec.Code != tt.wantCode {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantCode)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q", ct)
			}
			body := rec.Body.String()
			if !strings.Contains(body, tt.wantBody) || strings.Contains(body, "disk I/O") {
				t.Errorf("body = %s, want containing %s", body, tt.wantBody)
			}
			if !json.Valid(rec.Body.Bytes()) {
				t.Errorf("body is not valid JSON: %s", body)
			}
		})
	}
}
