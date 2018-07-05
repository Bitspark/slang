package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
	"encoding/json"
	"github.com/Bitspark/slang/pkg/utils"
)

var jsonWriteOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
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
	oFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}
			b, _ := json.Marshal(&i)
			out.Push(utils.Binary(b))
		}
	},
}
