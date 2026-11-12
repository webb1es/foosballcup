package server

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"time"
)

const shutdownTimeout = 10 * time.Second

// ReadinessCheck reports whether a dependency the server needs is reachable.
type ReadinessCheck func(ctx context.Context) error

type Server struct {
	http   *http.Server
	logger *slog.Logger
}

// Deps are the data sources the HTTP handlers read from.
type Deps struct {
	Ready        []ReadinessCheck
	LoadSnapshot SnapshotLoader
	UI           fs.FS
}

func New(addr string, logger *slog.Logger, deps Deps) *Server {
	return &Server{
		http: &http.Server{
			Addr:              addr,
			Handler:           routes(logger, deps),
			ReadHeaderTimeout: 5 * time.Second,
			IdleTimeout:       60 * time.Second,
			// No WriteTimeout: the SSE stream on /api/events keeps responses open indefinitely.
		},
		logger: logger,
	}
}

// Run serves until ctx is cancelled, then shuts down gracefully.
func (s *Server) Run(ctx context.Context) error {
	listenErr := make(chan error, 1)
	go func() {
		s.logger.Info("listening", "addr", s.http.Addr)
		err := s.http.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		listenErr <- err
	}()

	select {
	case err := <-listenErr:
		if err != nil {
			return fmt.Errorf("listen on %s: %w", s.http.Addr, err)
		}
		return nil
	case <-ctx.Done():
	}

	s.logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := s.http.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return <-listenErr
}

func routes(logger *slog.Logger, deps Deps) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /readyz", readyzHandler(logger, deps.Ready))
	mux.HandleFunc("GET /api/tournament", tournamentHandler(logger, deps.LoadSnapshot))
	mux.HandleFunc("GET /api/", apiNotFound)
	if deps.UI != nil {
		mux.Handle("GET /", uiHandler(deps.UI))
	}
	return mux
}
