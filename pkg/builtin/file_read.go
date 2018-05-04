package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
	"io/ioutil"
)

var fileReadOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		Services: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
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
		},
	},
	oFunc: func(srvs map[string]*core.Service, dels map[string]*core.Delegate, store interface{}) {
		in := srvs[core.MAIN_SERVICE].In()
		out := srvs[core.MAIN_SERVICE].Out()
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
