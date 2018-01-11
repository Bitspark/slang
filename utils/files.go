package utils

import (
	"strings"
	"os"
	"errors"
)

func FileWithFileEnding(filename string, fileEndings []string) (string, error) {
	for _, fileEnding := range fileEndings {
		if strings.HasSuffix(filename, fileEnding) {
			if _, err := os.Stat(filename); err == nil {
				return filename, nil
			} else {
				return "", err
			}
		}
	}

	for _, fileEnding := range fileEndings {
		filenameWithEnding := filename + fileEnding
		if _, err := os.Stat(filenameWithEnding); err == nil {
			return filenameWithEnding, nil
		}
	}

	return "", errors.New("no appropriate file found")
}