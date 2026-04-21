package auth

import "strings"

type StreamKeyValidator struct {
	// allowed is the single permitted key. Empty means any non-empty key is valid.
	allowed string
}

func NewStreamKeyValidator(allowed string) StreamKeyValidator {
	return StreamKeyValidator{allowed: allowed}
}

func (v StreamKeyValidator) ValidPath(path string) bool {
	trimmed := strings.TrimPrefix(path, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 || parts[0] != "live" || parts[1] == "" {
		return false
	}
	// Strip the _rtc suffix used by the FFmpeg-transcoded WebRTC copy.
	key := strings.TrimSuffix(parts[1], "_rtc")
	if key == "" {
		return false
	}
	// When no specific key is configured, accept any non-empty stream key.
	if v.allowed == "" {
		return true
	}
	return key == v.allowed
}
