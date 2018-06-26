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
)

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
		httpDef, err := api.ConstructHttpEndpoint(env, 12345, ri.Fqn, ri.Gens, ri.Props)
		if err != nil {
			data = outJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E0002"}}
			writeJSON(w, &data)
			return
		}

		bytes, _ := yaml.Marshal(httpDef)
		ioutil.WriteFile(
			filepath.Join(env.WorkingDir(), strings.Replace(ri.Fqn + "Packed", ".", string(filepath.Separator), -1) + ".yaml"),
			bytes,
			0644,
		)
	}},
}}
