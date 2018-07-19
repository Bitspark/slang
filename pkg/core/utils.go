package core

import (
	"os"
	"strings"
)

func IsMarker(item interface{}) bool {
	if _, ok := item.(BOS); ok {
		return true
	}
	if _, ok := item.(EOS); ok {
		return true
	}
	return false
}

func EnsureEnvironVar(key string, dfltVal string) {
	if val := os.Getenv(key); strings.Trim(val, " ") == "" {
		os.Setenv(key, dfltVal)
	}
}
