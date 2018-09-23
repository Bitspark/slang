package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, _ := reader.ReadString('\n')
		requestJson := line
		var request interface{}
		json.Unmarshal([]byte(requestJson), &request)
		response := {{ .Method }}(request)
		responseJson, _ := json.Marshal(response)
		fmt.Fprintln(os.Stdout, string(responseJson))
	}
}