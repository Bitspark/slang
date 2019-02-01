package daemon

import (
	"encoding/json"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"io/ioutil"
	"log"
	"net/http"
)

var DefinitionService = &Service{map[string]*Endpoint{
	"/": {func(st Storage, w http.ResponseWriter, r *http.Request) {
		type operatorDefJSON struct {
			Def  core.OperatorDef `json:"def"`
			Type string           `json:"type"`
		}

		type outJSON struct {
			Objects []operatorDefJSON `json:"objects"`
			Status  string            `json:"status"`
			Error   *Error            `json:"error,omitempty"`
		}

		var dataOut outJSON
		var opDefList []operatorDefJSON
		var err error

		opIds, err := st.List()

		if err == nil {
			builtinOpIds := elem.GetBuiltinIds()

			// Gather builtin/elementary opDefs
			for _, opId := range builtinOpIds {
				opDef, err := elem.GetOperatorDef(opId)

				if err != nil {
					break
				}

				opDefList = append(opDefList, operatorDefJSON{
					Type: "elementary",
					Def:  *opDef,
				})
			}

			if err == nil {
				// Gather opDefs from local & lib
				for _, opId := range opIds {
					opDef, err := st.Load(opId)
					if err != nil {
						continue
					}

					opType := "local"
					if st.IsLibrary(opId) {
						opType = "library"
					}

					opDefList = append(opDefList, operatorDefJSON{
						Type: opType,
						Def:  *opDef,
					})
				}
			}
		}

		if err == nil {
			dataOut = outJSON{Status: "success", Objects: opDefList}
		} else {
			dataOut = outJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E000X"}}
		}

		w.WriteHeader(200)
		err = writeJSON(w, dataOut)
		if err != nil {
			log.Print(err)
		}
	}},
	"/def/": {func(e Storage, w http.ResponseWriter, r *http.Request) {
		fail := func(err *Error) {
			sendFailure(w, &responseBad{err})
		}

		if r.Method == "POST" {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				fail(&Error{Msg: err.Error(), Code: "E000X"})
				return
			}

			var def core.OperatorDef
			err = json.Unmarshal(body, &def)
			if err != nil {
				fail(&Error{Msg: err.Error(), Code: "E000X"})
				return
			}

			_, err = e.Store(def)

			if err != nil {
				fail(&Error{Msg: err.Error(), Code: "E000X"})
				return
			}

			sendSuccess(w, nil)
		}
	}},
}}
