package auth

import (
	"encoding/json"
	"net/http"
)

type MediaMTXAuthRequest struct {
	Action string `json:"action"`
	Path   string `json:"path"`
}

type MediaMTXAuthHandler struct {
	validator StreamKeyValidator
}

func NewMediaMTXAuthHandler(validator StreamKeyValidator) MediaMTXAuthHandler {
	return MediaMTXAuthHandler{validator: validator}
}

func (h MediaMTXAuthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload MediaMTXAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if !h.validator.ValidPath(payload.Path) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// MVP policy: allow both publish and read only for a valid stream key path.
	switch payload.Action {
	case "publish", "read":
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "unsupported action", http.StatusForbidden)
	}
}
