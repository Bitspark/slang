package elem

import (
	"github.com/Bitspark/slang/pkg/core"
)

var constOpCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "trigger",
				},
				Out: core.TypeDef{
					Type:    "generic",
					Generic: "valueType",
				},
			},
		},
		PropertyDefs: map[string]*core.TypeDef{
			"value": {
				Type: "generic",
				Generic: "valueType",
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		v := op.Property("value")
		for !op.CheckStop() {
			if i := in.Pull(); !core.IsMarker(i) {
				out.Push(v)
			} else {
				out.Push(i)
			}
		}
	},
}
