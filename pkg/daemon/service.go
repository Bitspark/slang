package daemon

import (
	"net/http"
	"encoding/json"
	"io"
	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"strings"
)

type DaemonService struct {
	Routes map[string]*DaemonEndpoint
}

type DaemonEndpoint struct {
	Handle func(w http.ResponseWriter, r *http.Request)
}

func readJSON(r io.Reader, dat interface{}) error {
	dec := json.NewDecoder(r)
	return dec.Decode(&dat)
}

func writeJSON(w io.Writer, dat interface{}) error {
	return json.NewEncoder(w).Encode(dat)
}

type inJSON struct {
	WorkingDir       string                   `json:"workingdir"`
	OperatorFilePath string                   `json:"operator"`
	Generics         map[string]*core.TypeDef `json:"generics"`
	Properties       *core.Properties         `json:"properties"`
}

type outJSON struct {
	Objects []OperatorDefJSON `json:"objects"`
	Status  string            `json:"status"`
	Error   Error             `json:"error,omitempty"`
}

type OperatorDefJSON struct {
	Name string           `json:"name"`
	Def  core.OperatorDef `json:"def"`
	Type string           `json:"type"`
}

type Error struct {
	Msg  string `json:"msg"`
	Code string `json:"code"`
}

var OperatorDefService = &DaemonService{map[string]*DaemonEndpoint{
	/*
	   Get all OperatorDefs of all local, stdlib and elementaries

		REQUEST:

		{
			workingdir: path
		}


		RESPONSE:

	  	{
			objects: [
				{
					name: str,
	 				def: OperatorDef,
					type: str(local|elementary|lib),
				},
			]
	    }

	 */
	"/": {func(w http.ResponseWriter, r *http.Request) {
		var dataIn inJSON
		var dataOut outJSON
		var err error
		readJSON(r.Body, &dataIn)

		e := api.NewEnviron(dataIn.WorkingDir)

		opNames, err := e.ListOperatorNames()
		opDefList := make([]OperatorDefJSON, len(opNames))

		if err == nil {
			for i, opFQName := range opNames {

				opDefFilePath, err := e.GetOperatorDefFilePath(strings.Replace(opFQName, ".", ".", -1), "")

				if err == nil {
					break;
				}

				if opDef, err := e.ReadOperatorDef(opDefFilePath, nil); err != nil {

					opType := "lib"
					if e.IsLocalOperator(opDefFilePath) {
						opType = "local"
					}

					opDefList[i] = OperatorDefJSON{
						Name: opFQName,
						Type: opType,
						Def:  opDef,
					}
				}
			}
		}

		if err != nil {
			dataOut = outJSON{Status: "error", Error: Error{err.Error(), "E0001"}}
		} else {
			dataOut = outJSON{Status: "success", Objects: opDefList}
		}

		writeJSON(w, dataOut)
	}},
}}
