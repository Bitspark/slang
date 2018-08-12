package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"strings"
)

var stringSplitCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "string",
				},
				Out: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "string",
					},
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
		PropertyDefs: map[string]*core.TypeDef{
			"separator": {
				Type: "string",
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		sep := op.Property("separator").(string)
		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			str := i.(string)
			strs := strings.Split(str, sep)

			out.PushBOS()
			for _, s := range strs {
				out.Stream().Push(s)
			}
			out.PushEOS()
		}
	},
}
