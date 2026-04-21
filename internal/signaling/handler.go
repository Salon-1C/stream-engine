package signaling

import (
	"encoding/json"
	"net/http"
	"strings"

	"stream-engine/internal/auth"
	"stream-engine/internal/media"
)

type Handler struct {
	mediaBaseURL string
	validator    auth.StreamKeyValidator
	stats        *media.SessionStats
}

func New(mediaBaseURL string, validator auth.StreamKeyValidator, stats *media.SessionStats) Handler {
	return Handler{
		mediaBaseURL: strings.TrimSuffix(mediaBaseURL, "/"),
		validator:    validator,
		stats:        stats,
	}
}

type ViewerSessionResponse struct {
	WHEPURL string `json:"whep_url"`
}

func (h Handler) CreateViewerSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Query().Get("path")
	if !h.validator.ValidPath(path) {
		http.Error(w, "invalid stream path", http.StatusBadRequest)
		return
	}

	// Point WebRTC clients at the _rtc path where FFmpeg re-publishes the
	// stream with Opus audio (transcoded from RTMP AAC).
	rtcPath := strings.TrimPrefix(path, "/") + "_rtc"
	res := ViewerSessionResponse{
		WHEPURL: h.mediaBaseURL + "/" + rtcPath + "/whep",
	}

	writeJSON(w, res, http.StatusOK)
}

func (h Handler) AddViewer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.stats.AddViewer()
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) RemoveViewer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.stats.ViewerCount() > 0 {
		h.stats.RemoveViewer()
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) Stats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, map[string]int64{
		"viewers": h.stats.ViewerCount(),
	}, http.StatusOK)
}

func writeJSON(w http.ResponseWriter, payload any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
