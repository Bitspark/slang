package elem

import (
	"encoding/json"
	"github.com/Bitspark/slang/pkg/core"
)

var encodingJSONWriteCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type:    "generic",
					Generic: "itemType",
				},
				Out: core.TypeDef{
					Type: "binary",
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}
			b, err := json.Marshal(&i)
			if err != nil {
				panic(err)
			}
			out.Push(core.Binary(b))
		}
	},
}
