package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestUIHandler(t *testing.T) {
	ui := fstest.MapFS{
		"index.html":         {Data: []byte("<html>app</html>")},
		"logo.svg":           {Data: []byte("<svg/>")},
		"assets/app-abc1.js": {Data: []byte("console.log(1)")},
	}

	tests := []struct {
		name       string
		path       string
		wantCode   int
		wantBody   string
		wantCache  string
		wantPolicy bool
	}{
		{name: "root serves the app", path: "/", wantCode: http.StatusOK, wantBody: "app", wantCache: "no-cache", wantPolicy: true},
		{name: "client route serves the app", path: "/fixtures", wantCode: http.StatusOK, wantBody: "app", wantCache: "no-cache", wantPolicy: true},
		{name: "hashed asset is cached for a year", path: "/assets/app-abc1.js", wantCode: http.StatusOK, wantBody: "console.log", wantCache: "public, max-age=31536000, immutable"},
		{name: "public file is revalidated", path: "/logo.svg", wantCode: http.StatusOK, wantBody: "<svg/>", wantCache: "no-cache"},
		{name: "missing file is a 404, not the app", path: "/assets/missing.js", wantCode: http.StatusNotFound},
		{name: "unknown api path is a json 404", path: "/api/nope", wantCode: http.StatusNotFound, wantBody: `"error":"not found"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			routes(discardLogger(), Deps{UI: ui}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if rec.Code != tt.wantCode {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantCode)
			}
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("body = %q, want containing %q", rec.Body.String(), tt.wantBody)
			}
			if tt.wantCache != "" && rec.Header().Get("Cache-Control") != tt.wantCache {
				t.Errorf("Cache-Control = %q, want %q", rec.Header().Get("Cache-Control"), tt.wantCache)
			}
			if tt.wantPolicy && rec.Header().Get("Content-Security-Policy") == "" {
				t.Error("Content-Security-Policy header missing")
			}
		})
	}
}

func TestUIHandler_NotBuilt_Returns503(t *testing.T) {
	rec := httptest.NewRecorder()
	routes(discardLogger(), Deps{UI: fstest.MapFS{}}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
}
