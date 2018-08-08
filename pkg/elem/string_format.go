package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"fmt"
)

var stringFormatCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"format": {
							Type: "string",
						},
						"{variables}": {
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
			"variables": {
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
		vars := op.Property("variables").([]interface{})
		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			data := i.(map[string]interface{})
			format := data["format"].(string)
			var vals []interface{}
			for _, v := range vars {
				val := data[v.(string)]
				vals = append(vals, val)
			}

			out.Push(fmt.Sprintf(format, vals...))
		}
	},
}
