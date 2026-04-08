package server

import (
	"net/http"

	"stream-engine/internal/auth"
	"stream-engine/internal/config"
	"stream-engine/internal/media"
	"stream-engine/internal/signaling"
)

type Server struct {
	cfg       config.Config
	authz     auth.MediaMTXAuthHandler
	signaling signaling.Handler
}

func New(cfg config.Config) *Server {
	validator := auth.NewStreamKeyValidator(cfg.AllowedStreamKey)
	stats := media.NewSessionStats()

	return &Server{
		cfg:       cfg,
		authz:     auth.NewMediaMTXAuthHandler(validator),
		signaling: signaling.New(cfg.MediaMTXHTTPURL, validator, stats),
	}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	// MediaMTX authentication hook.
	mux.HandleFunc("/auth/mediamtx", s.authz.Handle)

	// Minimal signaling/support API for the demo page.
	mux.HandleFunc("/api/viewer-session", s.signaling.CreateViewerSession)
	mux.HandleFunc("/api/viewers/connect", s.signaling.AddViewer)
	mux.HandleFunc("/api/viewers/disconnect", s.signaling.RemoveViewer)
	mux.HandleFunc("/api/stats", s.signaling.Stats)

	// Static demo web UI.
	webFS := http.FileServer(http.Dir("web/demo"))
	mux.Handle("/", webFS)

	return withCORS(mux)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
