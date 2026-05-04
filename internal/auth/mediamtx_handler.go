package auth

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/golang-jwt/jwt/v5"
)

type MediaMTXAuthRequest struct {
	Action string `json:"action"`
	Path   string `json:"path"`
	Query  string `json:"query"` // MediaMTX forwards the client query string here
}

type MediaMTXAuthHandler struct {
	validator StreamKeyValidator
	jwtSecret []byte
}

func NewMediaMTXAuthHandler(validator StreamKeyValidator, jwtSecret string) MediaMTXAuthHandler {
	return MediaMTXAuthHandler{validator: validator, jwtSecret: []byte(jwtSecret)}
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

	// Path must still follow /live/<key> convention.
	if !h.validator.ValidPath(payload.Path) {
		http.Error(w, "forbidden: invalid path", http.StatusForbidden)
		return
	}

	// ---- JWT validation ------------------------------------------------
	// Token arrives as ?token=<jwt> in the client URL, forwarded by MediaMTX
	// via the "query" field of the auth webhook payload.
	values, _ := url.ParseQuery(payload.Query)
	tokenStr := values.Get("token")
	if tokenStr == "" {
		http.Error(w, "unauthorized: missing token", http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		// Reject anything that is not HS256 (matches Java JwtTokenAdapter).
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return h.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "unauthorized: invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "unauthorized: invalid claims", http.StatusUnauthorized)
		return
	}

	// ---- Role enforcement -----------------------------------------------
	// Claim name "roleCode" matches JwtTokenAdapter.CLAIM_ROLE in Java.
	role, _ := claims["roleCode"].(string)

	switch payload.Action {
	case "publish":
		// Only professors may broadcast.
		if role != "PROFESSOR" {
			http.Error(w, "forbidden: only professors can publish", http.StatusForbidden)
			return
		}
	case "read":
		// Any authenticated user may watch; token already validated above.
	default:
		http.Error(w, "forbidden: unsupported action", http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
}
