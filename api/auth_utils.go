package api

import (
	"strings"
)

func isHashedPassword(password string) bool {
	if len(password) < 60 {
		return false
	}
	return strings.HasPrefix(password, "$2a$") ||
		strings.HasPrefix(password, "$2b$") ||
		strings.HasPrefix(password, "$2y$")
}
