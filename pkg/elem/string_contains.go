package elem

import (
	"strings"

	"github.com/Bitspark/slang/pkg/core"
)

var stringContainsCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"str": {
							Type: "string",
						},
						"substr": {
							Type: "string",
						},
					},
				},
				Out: core.TypeDef{
					Type: "boolean",
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

			dataIn := i.(map[string]interface{})
			str := dataIn["str"].(string)
			subStr := dataIn["substr"].(string)
			dataOut := strings.Contains(str, subStr)

			out.Push(dataOut)
		}
	},
}
