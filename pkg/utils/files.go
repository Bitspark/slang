package utils

import (
	"fmt"
	"os"
	"strings"
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

	return "", fmt.Errorf("%s: no appropriate YAML or JSON file for given basename found", filename)
}

func IsJSON(opDefFilePath string) bool {
	return strings.HasSuffix(opDefFilePath, ".json")
}

func IsYAML(opDefFilePath string) bool {
	return strings.HasSuffix(opDefFilePath, ".yaml") || strings.HasSuffix(opDefFilePath, ".yml")
}
