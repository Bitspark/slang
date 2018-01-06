package builtin

import (
	"slang/core"
)

func isMarker(item interface{}) bool {
	if _, ok := item.(core.BOS); ok {
		return true
	}
	if _, ok := item.(core.EOS); ok {
		return true
	}
	return false
}
