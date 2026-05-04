package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"stream-engine/internal/auth"
	"stream-engine/internal/config"
	"stream-engine/internal/media"
	"stream-engine/internal/queue"
	"stream-engine/internal/signaling"
)

type Server struct {
	cfg       config.Config
	authz     auth.MediaMTXAuthHandler
	signaling signaling.Handler
	publisher *queue.Publisher
}

func New(cfg config.Config) *Server {
	validator := auth.NewStreamKeyValidator(cfg.AllowedStreamKey)
	stats := media.NewSessionStats()
	var publisher *queue.Publisher
	if cfg.RabbitMQURL != "" {
		pub, err := queue.NewPublisher(cfg.RabbitMQURL, cfg.RabbitMQQueue)
		if err != nil {
			log.Printf("rabbitmq publisher disabled: %v", err)
		} else {
			publisher = pub
			log.Printf("rabbitmq publisher enabled on queue %q", cfg.RabbitMQQueue)
		}
	}

	return &Server{
		cfg:       cfg,
		authz:     auth.NewMediaMTXAuthHandler(validator, cfg.JWTSecret),
		signaling: signaling.New(cfg.MediaMTXHTTPURL, validator, stats),
		publisher: publisher,
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
	mux.HandleFunc("/internal/recordings/segment-complete", s.handleSegmentComplete)

	// Static demo web UI.
	webFS := http.FileServer(http.Dir("web/demo"))
	mux.Handle("/", webFS)

	return withCORS(mux)
}

func (s *Server) handleSegmentComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.publisher == nil {
		http.Error(w, "queue not configured", http.StatusServiceUnavailable)
		return
	}
	var payload struct {
		StreamPath    string `json:"streamPath"`
		SegmentPath   string `json:"segmentPath"`
		ContentBase64 string `json:"contentBase64"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	payload.StreamPath = strings.TrimSpace(payload.StreamPath)
	payload.SegmentPath = strings.TrimSpace(payload.SegmentPath)
	payload.ContentBase64 = strings.TrimSpace(payload.ContentBase64)
	if payload.StreamPath == "" {
		http.Error(w, "streamPath is required", http.StatusBadRequest)
		return
	}
	if payload.SegmentPath == "" || payload.ContentBase64 == "" {
		http.Error(w, "segmentPath and contentBase64 are required", http.StatusBadRequest)
		return
	}
	err := s.publisher.Publish(r.Context(), queue.RecordingMessage{
		StreamPath:    payload.StreamPath,
		SegmentPath:   payload.SegmentPath,
		ContentBase64: payload.ContentBase64,
		Timestamp:     time.Now().UTC(),
	})
	if err != nil {
		http.Error(w, "failed to publish", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
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
