package builtin

import (
	"slang/core"
	"io/ioutil"
)

var fileReadOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		In: core.PortDef{
			Type: "string",
		},
		Out: core.PortDef{
			Type: "map",
			Map: map[string]*core.PortDef{
				"content": {
					Type: "binary",
				},
				"error": {
					Type: "string",
				},
			},
		},
	},
	oFunc: func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
		for true {
			file, marker := in.PullString()
			if marker != nil {
				out.Push(marker)
				continue
			}

			content, err := ioutil.ReadFile(file)
			if err != nil {
				out.Map("content").Push(nil)
				out.Map("error").Push(err.Error())
				continue
			}

			out.Map("content").Push(content)
			out.Map("error").Push(nil)
		}
	},
}
