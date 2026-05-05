package auth

import (
	"encoding/json"
	"log"
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
	log.Printf("[DEBUG] Auth request: method=%s", r.Method)
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload MediaMTXAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("[DEBUG] JSON decode error: %v", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	log.Printf("[DEBUG] Auth payload: action=%s path=%s", payload.Action, payload.Path)

	// Path must still follow /live/<key> convention.
	if !h.validator.ValidPath(payload.Path) {
		log.Printf("[DEBUG] Invalid path: %s", payload.Path)
		http.Error(w, "forbidden: invalid path", http.StatusForbidden)
		return
	}
	log.Printf("[DEBUG] Path validation passed")

	// ---- JWT validation ------------------------------------------------
	// Token arrives as ?token=<jwt> in the client URL, forwarded by MediaMTX
	// via the "query" field of the auth webhook payload.
	values, _ := url.ParseQuery(payload.Query)
	tokenStr := values.Get("token")
	log.Printf("[DEBUG] Query string: %s", payload.Query)
	if tokenStr == "" {
		http.Error(w, "unauthorized: missing token", http.StatusUnauthorized)
		return
	}
	log.Printf("[DEBUG] Token extracted (first 20 chars): %s...", tokenStr[:20])

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		// Reject anything that is not HS256 (matches Java JwtTokenAdapter).
		log.Printf("[DEBUG] Token algorithm: %s", t.Method.Alg())
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("[DEBUG] ERROR: Token is not HMAC, method type: %T", t.Method)
			return nil, jwt.ErrSignatureInvalid
		}
		return h.jwtSecret, nil
	})
	if err != nil {
		log.Printf("[DEBUG] JWT Parse error: %v", err)
		http.Error(w, "unauthorized: invalid token", http.StatusUnauthorized)
		return
	}
	if !token.Valid {
		log.Printf("[DEBUG] Token is invalid (Valid=false)")
		http.Error(w, "unauthorized: invalid token", http.StatusUnauthorized)
		return
	}
	log.Printf("[DEBUG] Token parsed successfully")

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Printf("[DEBUG] ERROR: claims are not MapClaims, type: %T", token.Claims)
		http.Error(w, "unauthorized: invalid claims", http.StatusUnauthorized)
		return
	}
	log.Printf("[DEBUG] Claims: %v", claims)

	// ---- Role enforcement -----------------------------------------------
	// Claim name "roleCode" matches JwtTokenAdapter.CLAIM_ROLE in Java.
	role, _ := claims["roleCode"].(string)
	log.Printf("[DEBUG] Role from token: %s", role)

	switch payload.Action {
	case "publish":
		// Only professors may broadcast.
		if role != "PROFESSOR" {
			log.Printf("[DEBUG] Publish rejected: role=%s (not PROFESSOR)", role)
			http.Error(w, "forbidden: only professors can publish", http.StatusForbidden)
			return
		}
		log.Printf("[DEBUG] Publish allowed for PROFESSOR")
	case "read":
		// Any authenticated user may watch; token already validated above.
		log.Printf("[DEBUG] Read access granted for role=%s", role)
	default:
		log.Printf("[DEBUG] Unknown action: %s", payload.Action)
		http.Error(w, "forbidden: unsupported action", http.StatusForbidden)
		return
	}

	log.Printf("[DEBUG] Auth successful for action=%s path=%s", payload.Action, payload.Path)
	w.WriteHeader(http.StatusOK)
}
