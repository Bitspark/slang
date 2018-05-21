package daemon

import (
	"net/http"
	"encoding/json"
	"io"
	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"strings"
	"github.com/Bitspark/slang/pkg/builtin"
	"path/filepath"
	"log"
	"github.com/Bitspark/slang/pkg/utils"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"os"
)

type DaemonService struct {
	Routes map[string]*DaemonEndpoint
}

type DaemonEndpoint struct {
	Handle func(w http.ResponseWriter, r *http.Request)
}

func writeJSON(w io.Writer, dat interface{}) error {
	return json.NewEncoder(w).Encode(dat)
}

type Error struct {
	Msg  string `json:"msg"`
	Code string `json:"code"`
}

var OperatorDefService = &DaemonService{map[string]*DaemonEndpoint{
	"/": {func(w http.ResponseWriter, r *http.Request) {
		type operatorDefJSON struct {
			Name string           `json:"name"`
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
		cwd := r.FormValue("cwd")

		e := api.NewEnviron(cwd)

		opNames, err := e.ListOperatorNames()

		if err == nil {
			builtinOpNames := builtin.GetBuiltinNames()

			// Gather builtin/elementary opDefs
			for _, opFQName := range builtinOpNames {
				opDef, err := builtin.GetOperatorDef(opFQName)

				if err != nil {
					break
				}

				opDefList = append(opDefList, operatorDefJSON{
					Name: opFQName,
					Type: "elementary",
					Def:  opDef,
				})
			}

			if err == nil {
				// Gather opDefs from local & lib
				for _, opFQName := range opNames {
					opDefFilePath, _, err := e.GetFilePathWithFileEnding(strings.Replace(opFQName, ".", string(filepath.Separator), -1), "")
					if err != nil {
						continue
					}

					opDef, err := e.ReadOperatorDef(opDefFilePath, nil)
					if err != nil {
						continue
					}

					opType := "lib"
					if e.IsLocalOperator(opFQName) {
						opType = "local"
					}

					opDefList = append(opDefList, operatorDefJSON{
						Name: opFQName,
						Type: opType,
						Def:  opDef,
					})
				}
			}
		}

		if err == nil {
			dataOut = outJSON{Status: "success", Objects: opDefList}
		} else {
			dataOut = outJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E0001"}}
		}

		w.WriteHeader(200)
		err = writeJSON(w, dataOut)
		if err != nil {
			log.Print(err)
		}
	}},
	"/def/": {func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			dataOut := struct {
				Status string `json:"status"`
				Error  *Error `json:"error,omitempty"`
			}{}

			send := func() {
				w.WriteHeader(200)
				err := writeJSON(w, dataOut)
				if err != nil {
					log.Print(err)
				}
			}

			cwd := r.FormValue("cwd")
			opFQName := r.FormValue("fqop")

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				dataOut.Status = "error"
				dataOut.Error = &Error{Msg: err.Error(), Code: "E0003"}
				send()
				return
			}

			fileEnding := ""
			var def core.OperatorDef
			err = json.Unmarshal(body, &def)
			if err == nil {
				fileEnding = "json"
			}

			err = yaml.Unmarshal(body, &def)
			if err == nil {
				fileEnding = "yaml"
			}

			if fileEnding == "" {
				dataOut.Status = "error"
				dataOut.Error = &Error{Msg: err.Error(), Code: "E0003"}
				send()
				return
			}

			relPath := strings.Replace(opFQName, ".", string(filepath.Separator), -1)
			absPath := filepath.Join(cwd, relPath+".yaml")
			err = os.MkdirAll(filepath.Dir(absPath), 0644)
			if err != nil {
				dataOut.Status = "error"
				dataOut.Error = &Error{Msg: err.Error(), Code: "E0003"}
				send()
				return
			}

			body, err = yaml.Marshal(&def)
			if err != nil {
				dataOut.Status = "error"
				dataOut.Error = &Error{Msg: err.Error(), Code: "E0007"}
				send()
				return
			}

			err = ioutil.WriteFile(absPath, body, 0644)
			if err != nil {
				dataOut.Status = "error"
				dataOut.Error = &Error{Msg: err.Error(), Code: "E0003"}
				send()
				return
			}

			dataOut.Status = "success"
			send()
		}
	}},
	"/meta/visual/": {func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			dataOut := struct {
				Data   interface{} `json:"data,omitempty"`
				Status string      `json:"status"`
				Error  *Error      `json:"error,omitempty"`
			}{}

			send := func() {
				w.WriteHeader(200)
				err := writeJSON(w, dataOut)
				if err != nil {
					log.Print(err)
				}
			}

			var err error
			var b []byte
			cwd := r.FormValue("cwd")
			opFQName := r.FormValue("fqop")

			e := api.NewEnviron(cwd)

			// Find the operator first
			relPath := strings.Replace(opFQName, ".", string(filepath.Separator), -1)
			absPath, p, err := e.GetFilePathWithFileEnding(relPath, "")
			if err != nil {
				dataOut.Status = "error"
				dataOut.Error = &Error{Msg: err.Error(), Code: "E0003"}
				send()
				return
			}

			// Then find the appropriate visual file and read it
			absPath, _, err = e.GetFilePathWithFileEnding(relPath+"_visual", p)
			b, err = ioutil.ReadFile(absPath)
			if err != nil {
				dataOut.Status = "error"
				dataOut.Error = &Error{Msg: err.Error(), Code: "E0003"}
				send()
				return
			}

			// Marshal
			if utils.IsJSON(absPath) {
				err = json.Unmarshal(b, &dataOut.Data)
			} else if utils.IsYAML(absPath) {
				err = yaml.Unmarshal(b, &dataOut.Data)
				dataOut.Data = utils.CleanValue(dataOut.Data)
			}

			if err != nil {
				dataOut.Status = "error"
				dataOut.Error = &Error{Msg: err.Error(), Code: "E0003"}
				send()
				return
			}

			// And send
			dataOut.Status = "success"
			send()
		} else if r.Method == "POST" {
			dataOut := struct {
				Status string `json:"status"`
				Error  *Error `json:"error,omitempty"`
			}{}

			send := func() {
				w.WriteHeader(200)
				err := writeJSON(w, dataOut)
				if err != nil {
					log.Print(err)
				}
			}

			cwd := r.FormValue("cwd")
			opFQName := r.FormValue("fqop")

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				dataOut.Status = "error"
				dataOut.Error = &Error{Msg: err.Error(), Code: "E0003"}
				send()
				return
			}

			var data interface{}
			err = json.Unmarshal(body, &data)
			if err != nil {
				err = yaml.Unmarshal(body, &data)
			}
			if err != nil {
				dataOut.Status = "error"
				dataOut.Error = &Error{Msg: err.Error(), Code: "E0003"}
				send()
				return
			}

			relPath := strings.Replace(opFQName, ".", string(filepath.Separator), -1)
			absPath := filepath.Join(cwd, relPath+"_visual.yaml")
			err = os.MkdirAll(filepath.Dir(absPath), 0644)
			if err != nil {
				dataOut.Status = "error"
				dataOut.Error = &Error{Msg: err.Error(), Code: "E0003"}
				send()
				return
			}

			body, err = yaml.Marshal(&data)
			if err != nil {
				dataOut.Status = "error"
				dataOut.Error = &Error{Msg: err.Error(), Code: "E0003"}
				send()
				return
			}

			err = ioutil.WriteFile(absPath, body, 0644)
			if err != nil {
				dataOut.Status = "error"
				dataOut.Error = &Error{Msg: err.Error(), Code: "E0003"}
				send()
				return
			}

			dataOut.Status = "success"
			send()
		}
	}},
}}
