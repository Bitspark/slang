package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var dataValueId = uuid.MustParse("8b62495a-e482-4a3e-8020-0ab8a350ad2d")
var dataValueCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: dataValueId,
		Meta: core.BlueprintMetaDef{
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
		PropertyDefs: core.PropertyMap{
			"value": {
				core.TypeDef{
					Type:    "generic",
					Generic: "valueType",
				},
				nil,
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
