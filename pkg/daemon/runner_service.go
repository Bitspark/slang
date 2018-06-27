package daemon

import (
	"net/http"
	"encoding/json"
	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"io/ioutil"
	"path/filepath"
	"strings"
	"gopkg.in/yaml.v2"
	"strconv"
)

var port = 12345

var RunnerService = &DaemonService{map[string]*DaemonEndpoint{
	"/": {func(w http.ResponseWriter, r *http.Request) {
		type runInstructionJSON struct {
			Cwd   string          `json:"cwd"`
			Fqn   string          `json:"fqn"`
			Props core.Properties `json:"props"`
			Gens  core.Generics   `json:"gens"`
		}

		type outJSON struct {
			URL    string `json:"url,omitempty"`
			Status string `json:"status"`
			Error  *Error `json:"error,omitempty"`
		}

		var data outJSON

		decoder := json.NewDecoder(r.Body)
		var ri runInstructionJSON
		err := decoder.Decode(&ri)
		if err != nil {
			data = outJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E0001"}}
			writeJSON(w, &data)
			return
		}

		env := api.NewEnviron(ri.Cwd)
		httpDef, err := api.ConstructHttpEndpoint(env, port, ri.Fqn, ri.Gens, ri.Props)
		if err != nil {
			data = outJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E0002"}}
			writeJSON(w, &data)
			return
		}

		packagedOperator := strings.Replace(ri.Fqn + "Packed", ".", string(filepath.Separator), -1) + ".yaml"

		bytes, _ := yaml.Marshal(httpDef)
		ioutil.WriteFile(
			filepath.Join(env.WorkingDir(), packagedOperator),
			bytes,
			0644,
		)

		op, err := env.BuildAndCompileOperator(packagedOperator, nil, nil)
		if err != nil {
			data = outJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E0003"}}
			writeJSON(w, &data)
			return
		}

		op.Start()
		op.Main().In().Push(nil) // Start server

		data.Status = "success"
		data.URL = "http://localhost:" + strconv.Itoa(port)

		port++

		writeJSON(w, &data)
	}},
}}
