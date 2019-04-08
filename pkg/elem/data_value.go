package elem

import (
	"github.com/Bitspark/slang/pkg/core"
)

var dataValueId = "8b62495a-e482-4a3e-8020-0ab8a350ad2d"
var dataValueCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: dataValueId,
		Meta: core.OperatorMetaDef{
			Name:             "value",
			ShortDescription: "emits a constant value for each item",
			Icon:             "box-full",
			Tags:             []string{"data"},
			DocURL:           "https://bitspark.de/slang/docs/operator/value",
		},
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
				Type:    "generic",
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
