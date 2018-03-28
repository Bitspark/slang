package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
)

var plotOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
					},
				},
				Out: core.TypeDef{
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
	},
	oFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for {
			// TODO: Implement
			out.Push(in.Pull())
		}
	},
}
