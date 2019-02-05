package core

import (
	"encoding/base64"
	"errors"
	"strings"
)

func (b *Binary) UnmarshalJSON(bytes []byte) error {
	base64String := string(bytes)
	if !strings.HasPrefix(base64String, "\"base64:") || !strings.HasSuffix(base64String, "\"") {
		return errors.New("wrongly encoded base64")
	}

	binary, err := base64.StdEncoding.DecodeString(string(bytes)[8 : len(bytes)-1])
	if err != nil {
		*b = Binary(binary)
	}

	return err
}

func (b Binary) MarshalJSON() ([]byte, error) {
	return []byte("\"base64:" + base64.StdEncoding.EncodeToString(b) + "\""), nil
}
