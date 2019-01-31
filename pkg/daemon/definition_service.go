package daemon

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"github.com/Bitspark/slang/pkg/utils"
	"gopkg.in/yaml.v2"
)

const SuffixVisual = "_visual"

var uuidMap map[uuid.UUID]string

var DefinitionService = &Service{map[string]*Endpoint{
	"/": {func(e *api.Environ, w http.ResponseWriter, r *http.Request) {
		type operatorDefJSON struct {
			Id   string           `json:"id"`
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

		opIds, err := e.ListOperatorNames()

		if uuidMap == nil {
			uuidMap = make(map[uuid.UUID]string)
		}

		if err == nil {
			builtinOpIds := elem.GetBuiltinIds()

			// Gather builtin/elementary opDefs
			for _, opId := range builtinOpIds {
				opDef, err := elem.GetOperatorDef(opId)

				if err != nil {
					fmt.Println("[ERROR1]", err)
					break
				}

				opDefList = append(opDefList, operatorDefJSON{
					Id:   opId,
					Type: "elementary",
					Def:  opDef,
				})
			}

			if err == nil {
				// Gather opDefs from local & lib
				for _, opId := range opIds {
					opDefFilePath, _, err := e.GetFilePathWithFileEnding(strings.Replace(opId, ".", string(filepath.Separator), -1), "")

					if err != nil {
						fmt.Println("[ERROR2]", err)
						continue
					}

					opDef, err := e.ReadOperatorDef(opDefFilePath, nil)
					if err != nil {
						fmt.Println("[ERROR2]", err)
						continue
					}

					opType := "library"
					if e.IsLocalOperator(opId) {
						opType = "local"
					}

					opDefList = append(opDefList, operatorDefJSON{
						Id:   opId,
						Type: opType,
						Def:  opDef,
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
	"/def/": {func(e *api.Environ, w http.ResponseWriter, r *http.Request) {
		fail := func(err *Error) {
			sendFailure(w, &responseBad{err})
		}

		if r.Method == "POST" {
			/*
			 * POST OperatorDef
			 */
			cwd := e.WorkingDir()
			opId := r.FormValue("id")

			/* CHECK UUID is valid
			if !checkOperatorNameIsValid(opId) {
				fail(&Error{Msg: fmt.Sprintf("operator must start with capital letter and may only contain alphanumerics"), Code: "E000X"})
				return
			}
			*/

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

			relPath := strings.Replace(opId, ".", string(filepath.Separator), -1)
			absPath := filepath.Join(cwd, relPath+".yaml")
			_, err = EnsureDirExists(filepath.Dir(absPath))
			if err != nil {
				fail(&Error{Msg: err.Error(), Code: "E000X"})
				return
			}

			body, err = yaml.Marshal(&def)
			if err != nil {
				fail(&Error{Msg: err.Error(), Code: "E000X"})
				return
			}

			err = ioutil.WriteFile(absPath, body, os.ModePerm)
			if err != nil {
				fail(&Error{Msg: err.Error(), Code: "E000X"})
				return
			}
			sendSuccess(w, nil)
		}
	}},
	"/meta/visual/": {func(e *api.Environ, w http.ResponseWriter, r *http.Request) {
		fail := func(err *Error) {
			sendFailure(w, &responseBad{err})
		}
		/*
		 * GET meta/visual
		 */
		if r.Method == "GET" {
			var err error
			var b []byte
			opFQName := r.FormValue("fqop")

			// Find the operator first
			relPath := strings.Replace(opFQName, ".", string(filepath.Separator), -1)
			absPath, p, err := e.GetFilePathWithFileEnding(relPath, "")
			if err != nil {
				fail(&Error{Msg: err.Error(), Code: "E000X"})
				return
			}

			// Then find the appropriate visual file and read it
			absPath, _, err = e.GetFilePathWithFileEnding(relPath+SuffixVisual, p)
			b, err = ioutil.ReadFile(absPath)
			if err != nil {
				fail(&Error{Msg: err.Error(), Code: "E000X"})
				return
			}

			// Marshal
			resp := &responseOK{}
			if utils.IsJSON(absPath) {
				err = json.Unmarshal(b, &resp.Data)
			} else if utils.IsYAML(absPath) {
				err = yaml.Unmarshal(b, &resp.Data)
				resp.Data = utils.CleanValue(resp.Data)
			}

			if err != nil {
				fail(&Error{Msg: err.Error(), Code: "E000X"})
				return
			}
			// And send
			sendSuccess(w, resp)
		} else if r.Method == "POST" {
			/*
			 * POST meta/visual
			 */
			cwd := e.WorkingDir()
			opFQName := r.FormValue("fqop")

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				fail(&Error{Msg: err.Error(), Code: "E000X"})
				return
			}

			var data interface{}
			err = json.Unmarshal(body, &data)
			if err != nil {
				err = yaml.Unmarshal(body, &data)
			}
			if err != nil {
				fail(&Error{Msg: err.Error(), Code: "E000X"})
				return
			}

			relPath := strings.Replace(opFQName, ".", string(filepath.Separator), -1)
			absPath := filepath.Join(cwd, relPath+SuffixVisual+".yaml")
			_, err = EnsureDirExists(filepath.Dir(absPath))
			if err != nil {
				fail(&Error{Msg: err.Error(), Code: "E000X"})
				return
			}

			body, err = yaml.Marshal(&data)
			if err != nil {
				fail(&Error{Msg: err.Error(), Code: "E000X"})
				return
			}

			err = ioutil.WriteFile(absPath, body, os.ModePerm)
			if err != nil {
				fail(&Error{Msg: err.Error(), Code: "E000X"})
				return
			}
			sendSuccess(w, nil)
		}
	}},
}}
