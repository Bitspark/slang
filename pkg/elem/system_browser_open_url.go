package elem

import (
	"github.com/Bitspark/browser"
	"github.com/Bitspark/slang/pkg/core"
)

var systemBrowserOpenURLCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"url": {
							Type: "string",
						},
					},
				},
				Out: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"err": {
							Type: "string",
						},
					},
				},
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()

		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) || i == nil {
				out.Push(i)
				continue
			}

			im := i.(map[string]interface{})
			path := im["url"].(string)

			err := browser.OpenURL(path)

			if err != nil {
				out.Push(err.Error())
			} else {
				out.Push(nil)
			}
		}
	},
}
