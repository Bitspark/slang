package core

func IsMarker(item interface{}) bool {
	if _, ok := item.(BOS); ok {
		return true
	}
	if _, ok := item.(EOS); ok {
		return true
	}
	return false
}
