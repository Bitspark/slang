package elem

import (
	"strings"

	"github.com/Bitspark/slang/pkg/core"
)

var stringBeginswithCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: "9f274995-2726-4513-ac7c-f15ac7b68720",
		Meta: core.OperatorMetaDef{
			Name: "begins with",
		},
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
			dataOut := strings.HasPrefix(str, subStr)

			out.Push(dataOut)
		}
	},
}
