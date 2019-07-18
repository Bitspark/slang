package daemon

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
)

var DefinitionService = &Service{map[string]*Endpoint{
	"/": {func(w http.ResponseWriter, r *http.Request) {
		st := GetStorage(r)
		type blueprintJSON struct {
			Def  core.Blueprint `json:"def"`
			Type string         `json:"type"`
		}

		type outJSON struct {
			Objects []blueprintJSON `json:"objects"`
			Status  string          `json:"status"`
			Error   *Error          `json:"error,omitempty"`
		}

		var dataOut outJSON
		var err error
		blueprints := make([]blueprintJSON, 0)

		opIds, err := st.List()

		if err == nil {
			builtinOpIds := elem.GetBuiltinIds()

			// Gather builtin/elementary blueprints
			for _, opId := range builtinOpIds {
				blueprint, err := elem.GetBlueprint(opId)

				if err != nil {
					break
				}

				blueprints = append(blueprints, blueprintJSON{
					Type: "elementary",
					Def:  *blueprint,
				})
			}

			if err == nil {
				// Gather blueprints from local & lib
				for _, opId := range opIds {
					blueprint, err := st.Load(opId)
					if err != nil {
						continue
					}

					opType := "library"
					if st.IsSavedInWritableBackend(opId) {
						opType = "local"
					}

					blueprints = append(blueprints, blueprintJSON{
						Type: opType,
						Def:  *blueprint,
					})
				}
			}
		}

		if err == nil {
			dataOut = outJSON{Status: "success", Objects: blueprints}
		} else {
			dataOut = outJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E000X"}}
		}

		w.WriteHeader(200)
		err = writeJSON(w, dataOut)
		if err != nil {
			log.Print(err)
		}
	}},
	"/def/": {func(w http.ResponseWriter, r *http.Request) {
		st := GetStorage(r)
		fail := func(err *Error) {
			sendFailure(w, &responseBad{err})
		}

		if r.Method == "POST" {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				fail(&Error{Msg: err.Error(), Code: "E000X"})
				return
			}

			var def core.Blueprint
			err = json.Unmarshal(body, &def)
			if err != nil {
				fail(&Error{Msg: err.Error(), Code: "E000X"})
				return
			}

			_, err = st.Save(def)

			if err != nil {
				fail(&Error{Msg: err.Error(), Code: "E000X"})
				return
			}

			sendSuccess(w, nil)
		}
	}},
}}
