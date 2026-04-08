package auth

import "strings"

type StreamKeyValidator struct {
	allowed string
}

func NewStreamKeyValidator(allowed string) StreamKeyValidator {
	return StreamKeyValidator{allowed: allowed}
}

func (v StreamKeyValidator) ValidPath(path string) bool {
	trimmed := strings.TrimPrefix(path, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 {
		return false
	}

	return parts[0] == "live" && parts[1] == v.allowed
}
