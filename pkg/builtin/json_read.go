package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
	"encoding/json"
	"github.com/Bitspark/slang/pkg/utils"
)

var jsonReadOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "binary",
				},
				Out: core.TypeDef{
					Type:    "generic",
					Generic: "itemType",
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
	},
	oFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}
			var obj interface{}
			err := json.Unmarshal([]byte(i.(utils.Binary)), &obj)
			if err != nil {
				out.Push(nil)
				continue
			}
			obj = utils.CleanValue(obj)
			out.Push(obj) // TODO: Make this safer
		}
	},
}
