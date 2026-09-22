package utils

import (
	"fmt"

	"github.com/thoas/go-funk"
)

func AskForConfirmation(question string) bool {
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		fmt.Printf("%s [y/n]\n", question)
		return AskForConfirmation(question)
	}
	okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
	nokayResponses := []string{"n", "N", "no", "No", "NO"}
	if funk.Contains(okayResponses, response) {
		return true
	} else if funk.Contains(nokayResponses, response) {
		return false
	} else {
		fmt.Printf("%s [y/n]\n", question)
		return AskForConfirmation(question)
	}
}
