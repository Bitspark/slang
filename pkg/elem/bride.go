package elem

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/utils"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

type bridgeCfg struct {
	name    string
	ending  string
	sep     string
	isolate bool
	cmd     string
	args    []string
}

func createBridgeEnv(dir string, isolate bool, ipc, ending, sep, source, method string) (string, error) {
	opDir := "ops"
	tmp := "tmp" + fmt.Sprintf("%x", rand.Int63())

	var srcFile, opFile, hostFile string
	if isolate {
		opFile = "op"
		srcFile = filepath.Join(tmp, opFile+"."+ending)
		hostFile = filepath.Join(tmp, "host."+ending)

		os.MkdirAll(filepath.Join(dir, opDir, tmp), os.ModePerm)
	} else {
		opFile = tmp + sep + "op"
		srcFile = opFile + "." + ending
		hostFile = tmp + sep + "host." + ending

		os.MkdirAll(filepath.Join(dir, opDir), os.ModePerm)
	}

	err := ioutil.WriteFile(filepath.Join(dir, opDir, srcFile), []byte(source), os.ModePerm)
	if err != nil {
		return "", err
	}

	type ctx struct {
		OpFile string
		Method string
	}

	hostSrc, err := ioutil.ReadFile(filepath.Join(dir, ipc+"."+ending))
	if err != nil {
		return "", err
	}
	hostTpl, err := template.New("http").Parse(string(hostSrc))
	if err != nil {
		return "", err
	}

	srcBuf := new(bytes.Buffer)
	hostTpl.Execute(srcBuf, ctx{opFile, method})

	err = ioutil.WriteFile(filepath.Join(dir, opDir, hostFile), srcBuf.Bytes(), os.ModePerm)
	if err != nil {
		return "", err
	}

	if isolate {
		return filepath.Dir(filepath.Join(dir, opDir, hostFile)), nil
	} else {
		return filepath.Join(dir, opDir, hostFile), nil
	}
}

func createBridgeCfg(bridge bridgeCfg) *builtinConfig {
	return &builtinConfig{
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
			DelegateDefs: map[string]*core.DelegateDef{/*...*/},
			PropertyDefs: map[string]*core.TypeDef{
				"source": {
					Type: "string",
				},
				"method": {
					Type: "string",
				},
				"ipc": {
					Type: "string",
				},
				/*"delegates": {
					Type: "string",
				},*/
			},
		},
		opFunc: func(op *core.Operator) {
			ipc := op.Property("ipc").(string)
			source := op.Property("source").(string)
			method := op.Property("method").(string)

			dir := "./bridge/" + bridge.name + "/"

			file, err := createBridgeEnv(
				dir,
				bridge.isolate,
				ipc,
				bridge.ending, bridge.sep,
				source, method)
			if err != nil {
				panic(err)
			}

			args := bridge.args
			if !bridge.isolate {
				args = append(args, file)
			}
			cmd := exec.Command(bridge.cmd, args...)

			if bridge.isolate {
				cmd.Dir = file
			}
			cmd.Stderr = os.Stderr

			stdout, err := cmd.StdoutPipe()
			if err != nil {
				panic(err)
			}

			stdin, err := cmd.StdinPipe()
			if err != nil {
				panic(err)
			}

			err = cmd.Start()
			if err != nil {
				panic(err)
			}

			r := bufio.NewReader(stdout)

			in := op.Main().In()
			out := op.Main().Out()

			if ipc == "http" {
				url, _, err := r.ReadLine()
				if err != nil {
					panic(err)
				}

				client := &http.Client{}

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
			} else if ipc == "pipe" {
				for !op.CheckStop() {
					item := in.Pull()
					if core.IsMarker(item) {
						out.Push(item)
						continue
					}

					requestJson, err := json.Marshal(item)
					if err != nil {
						panic(err)
					}

					stdin.Write(requestJson)
					stdin.Write([]byte("\r\n"))

					responseJson, _, err := r.ReadLine()
					if err != nil {
						panic(err)
					}

					var result interface{}
					json.Unmarshal(responseJson, &result)
					result = utils.CleanValue(result)

					out.Push(result)
				}
			}
		},
	}
}
