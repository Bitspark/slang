package elem

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/utils"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func createPythonEnv(prefix, dir, source, method string) (*exec.Cmd, error) {
	os.MkdirAll(dir, os.ModePerm)

	err := ioutil.WriteFile(filepath.Join(dir, prefix+"_operator.py"), []byte(source), os.ModePerm)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(filepath.Join(dir, prefix+"_rpc_server.py"), []byte(
		`import json, sys
from BaseHTTPServer import HTTPServer, BaseHTTPRequestHandler
from `+prefix+`_operator import `+method+`

class OperatorServer(BaseHTTPRequestHandler):
    def _set_headers(self):
        self.send_response(200)
        self.send_header('Content-type', 'text/json')
        self.end_headers()

    def do_POST(self):
        request_body = self.rfile.read(int(self.headers.getheader('Content-Length')))
        response_body = json.dumps(`+method+`(json.loads(request_body)))

        self._set_headers()
        self.wfile.write(response_body)

if __name__ == '__main__':
    httpd = HTTPServer(("", 0), OperatorServer)
    port = httpd.socket.getsockname()[1]
    print "http://{}:{}".format("localhost", port)
    sys.stdout.flush()
    httpd.serve_forever()
`), os.ModePerm)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("python", filepath.Join(dir, prefix+"_rpc_server.py"))

	return cmd, nil
}

var bridgePythonCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type:    "generic",
					Generic: "inType",
				},
				Out: core.TypeDef{
					Type:    "generic",
					Generic: "outType",
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
		PropertyDefs: map[string]*core.TypeDef{
			"source": {
				Type: "string",
			},
			"method": {
				Type: "string",
			},
		},
	},
	opFunc: func(op *core.Operator) {
		source := op.Property("source").(string)
		method := op.Property("method").(string)

		cmd, err := createPythonEnv(
			strings.Replace(strings.ToLower(op.Name()), "#", "_", -1),
			"./python_scripts/",
			source, method)
		if err != nil {
			panic(err)
		}

		cmd.Stderr = os.Stderr

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			panic(err)
		}

		err = cmd.Start()
		if err != nil {
			panic(err)
		}

		r := bufio.NewReader(stdout)

		url, _, err := r.ReadLine()
		if err != nil {
			panic(err)
		}

		client := &http.Client{}

		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			item := in.Pull()
			if core.IsMarker(item) {
				out.Push(item)
				continue
			}

			requestBody, err := json.Marshal(item)
			if err != nil {
				panic(err)
			}

			req, err := http.NewRequest("POST", string(url), bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}

			fmt.Println("response Status:", resp.Status)
			fmt.Println("response Headers:", resp.Header)

			responseBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				resp.Body.Close()
				panic(err)
			}
			resp.Body.Close()

			fmt.Println("response Body:", string(responseBody))

			var result interface{}
			json.Unmarshal(responseBody, &result)
			result = utils.CleanValue(result)

			fmt.Println("Python responded: ", result)

			out.Push(result)
		}
	},
}
