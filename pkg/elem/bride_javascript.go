package elem

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/utils"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func createJavaScriptEnv(prefix, dir, source, method string) (*exec.Cmd, error) {
	os.MkdirAll(dir, os.ModePerm)

	err := ioutil.WriteFile(filepath.Join(dir, prefix+"-operator.js"), []byte(source), os.ModePerm)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(filepath.Join(dir, prefix+"-server.js"), []byte(
		`var http = require('http');
var op = require('./`+prefix+`-operator.js');

var server  = http.createServer(function (req, res) {
    var requestBody = '';
  	req.on('data', function (data) {
        requestBody += data;
    });
    req.on('end', function () {
        res.writeHead(200, {'Content-Type': 'text/json'});
        var responseBody = JSON.stringify(op.`+method+`(JSON.parse(requestBody)));
        res.write(responseBody);
        res.end();
    });
})
server.listen(0)
server.on('listening', function() {
  console.log('http://localhost:' + server.address().port)
})`), os.ModePerm)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("node", filepath.Join(dir, prefix+"-server.js"))

	return cmd, nil
}

var bridgeJavaScriptCfg = &builtinConfig{
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

		cmd, err := createJavaScriptEnv(
			strings.Replace(strings.ToLower(op.Name()), "#", "-", -1),
			"./javascript_scripts/",
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

			responseBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				resp.Body.Close()
				panic(err)
			}
			resp.Body.Close()

			var result interface{}
			json.Unmarshal(responseBody, &result)
			result = utils.CleanValue(result)

			out.Push(result)
		}
	},
}
