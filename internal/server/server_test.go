package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestRoutes(t *testing.T) {
	passing := func(context.Context) error { return nil }
	failing := func(context.Context) error { return errors.New("database unreachable") }

	tests := []struct {
		name     string
		method   string
		path     string
		checks   []ReadinessCheck
		wantCode int
		wantBody string
	}{
		{name: "healthz ok", method: http.MethodGet, path: "/healthz", wantCode: http.StatusOK, wantBody: "ok\n"},
		{name: "healthz rejects post", method: http.MethodPost, path: "/healthz", wantCode: http.StatusMethodNotAllowed},
		{name: "readyz with no checks", method: http.MethodGet, path: "/readyz", wantCode: http.StatusOK, wantBody: "ready\n"},
		{name: "readyz with passing checks", method: http.MethodGet, path: "/readyz", checks: []ReadinessCheck{passing, passing}, wantCode: http.StatusOK, wantBody: "ready\n"},
		{name: "readyz with a failing check", method: http.MethodGet, path: "/readyz", checks: []ReadinessCheck{passing, failing}, wantCode: http.StatusServiceUnavailable, wantBody: "not ready\n"},
		{name: "unknown path", method: http.MethodGet, path: "/nope", wantCode: http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			routes(discardLogger(), tt.checks).ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))

			if rec.Code != tt.wantCode {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantCode)
			}
			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Errorf("body = %q, want %q", rec.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestRun_ContextCancelled_ReturnsNil(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := New("127.0.0.1:0", discardLogger()).Run(ctx); err != nil {
		t.Errorf("Run() error = %v, want nil", err)
	}
}

func TestRun_PortInUse_ReturnsError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	defer ln.Close()

	if err := New(ln.Addr().String(), discardLogger()).Run(context.Background()); err == nil {
		t.Error("Run() error = nil, want listen error")
	}
}
