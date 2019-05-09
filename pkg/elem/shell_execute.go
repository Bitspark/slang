package elem

import (
	"os/exec"

	"github.com/Bitspark/slang/pkg/core"
)

var shellExecuteCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: "13cbad40-da00-40d7-bdcd-981b14ec346b",
		Meta: core.OperatorMetaDef{
			Name:             "shell execute",
			ShortDescription: "executes a shell command on the host system",
			Icon:             "terminal",
			Tags:             []string{"system"},
			DocURL:           "https://bitspark.de/slang/docs/operator/shell-execute",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"command": {
							Type: "string",
						},
						"arguments": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type: "string",
							},
						},
					},
				},
				Out: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"stdout": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type: "binary",
							},
						},
						"stderr": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type: "binary",
							},
						},
						"stdin": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type: "binary",
							},
						},
						"code": {
							Type: "string",
						},
					},
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{
			"user": {
				In: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "binary",
					},
				},
				Out: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"stdout": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type: "binary",
							},
						},
						"stderr": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type: "binary",
							},
						},
					},
				},
			},
		},
		PropertyDefs: map[string]*core.TypeDef{
			"bufferSize": {
				Type: "number",
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		user := op.Delegate("user")
		buffersize := int(op.Property("bufferSize").(float64))
		for !op.CheckStop() {
			i := in.Map("command").Pull()
			if core.IsMarker(i) || i == nil {
				in.Map("stdin").Pull()
				out.Push(i)
				continue
			}

			cmd := i.(string)
			args := []string{}
			for _, arg := range in.Map("arguments").Pull().([]interface{}) {
				args = append(args, arg.(string))
			}

			c := exec.Command(cmd, args...)

			stdout, _ := c.StdoutPipe()
			stderr, _ := c.StderrPipe()
			stdin, _ := c.StdinPipe()

			c.Start()
			// Redirect stdout to out port
			go func() {
				user.Out().Map("stdout").PushBOS()
				out.Map("stdout").PushBOS()
				bytes := make([]byte, buffersize)
				for {
					read, err := stdout.Read(bytes[:])
					if err != nil {
						break
					}
					chunk := core.Binary(bytes[0:read])
					user.Out().Map("stdout").Stream().Push(chunk)
					out.Map("stdout").Stream().Push(chunk)
				}
				user.Out().Map("stdout").PushEOS()
				out.Map("stdout").PushEOS()
			}()
			// Redirect stderr to out port
			go func() {
				user.Out().Map("stderr").PushBOS()
				out.Map("stderr").PushBOS()
				bytes := make([]byte, buffersize)
				for {
					read, err := stderr.Read(bytes[:])
					if err != nil {
						break
					}
					chunk := core.Binary(bytes[0:read])
					user.Out().Map("stderr").Stream().Push(chunk)
					out.Map("stderr").Stream().Push(chunk)
				}
				user.Out().Map("stderr").PushEOS()
				out.Map("stderr").PushEOS()
			}()
			// Redirect stdin to program and out port
			go func() {
				user.In().PullBOS()
				out.Map("stdin").PushBOS()
				for {
					i := user.In().Stream().Pull()
					if user.In().OwnEOS(i) {
						out.Map("stdin").PushEOS()
						stdin.Close()
						break
					}
					input := i.(core.Binary)
					stdin.Write(input)
					out.Map("stdin").Stream().Push(input)
				}
			}()
			err := c.Wait()
			if err != nil {
				out.Map("code").Push(err.Error())
			} else {
				out.Map("code").Push(nil)
			}
		}
	},
}
