package utils

import "encoding/base64"

func (b Binary) MarshalJSON() ([]byte, error) {
	return []byte("\"base64:" + base64.StdEncoding.EncodeToString(b) + "\""), nil
}
