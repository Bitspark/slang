package core

func IsEOS(item interface{}) bool {
	_, ok := item.(EOS)
	return ok
}

func IsBOS(item interface{}) bool {
	_, ok := item.(BOS)
	return ok
}

func IsMarker(item interface{}) bool {
	return IsBOS(item) || IsEOS(item)
}
