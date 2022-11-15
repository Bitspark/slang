package daemon

import (
	"net/http"

	"github.com/Bitspark/go-funk"
)

var RunOperatorService = &Service{map[string]*Endpoint{
	/*
	 *	Get all running operators
	 */
	"/": {func(w http.ResponseWriter, r *http.Request) {

		type outJSON struct {
			Objects []runningOperator `json:"objects"`
			Status  string            `json:"status"`
			Error   *Error            `json:"error,omitempty"`
		}

		var data outJSON

		if r.Method == "GET" {
			data = outJSON{Objects: funk.Values(runningOperatorManager.ops).([]runningOperator), Status: "success", Error: nil}
			writeJSON(w, &data)
		}
	}},
}}
