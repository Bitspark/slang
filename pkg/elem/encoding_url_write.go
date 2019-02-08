package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"net/url"
	"fmt"
)

var encodingURLWriteCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Meta: core.OperatorMetaDef{
			Name: "encode URL",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"{params}": {
							Type: "primitive",
						},
					},
				},
				Out: core.TypeDef{
					Type: "string",
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
		PropertyDefs: map[string]*core.TypeDef{
			"params": {
				Type: "stream",
				Stream: &core.TypeDef{
					Type: "string",
				},
			},
		},
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
			vals := url.Values{}
			im := i.(map[string]interface{})
			for key, value := range im {
				vals.Set(key, fmt.Sprintf("%v", value))
			}
			out.Push(vals.Encode())
		}
	},
}
