package daemon

import "net/http"

var RunnerService = &DaemonService{map[string]*DaemonEndpoint{
	"/": {func(w http.ResponseWriter, r *http.Request) {
	}},
}}