package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var dataUUIDCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: "a83bf9b2-cf1b-4b14-94c2-ea04d5cf70c0",
		Meta: core.OperatorMetaDef{
			Name: "generate UUID",
			ShortDescription: "generates a random UUID",
			Icon: "barcode-alt",
			Tags: []string{"data", "random"},
			DocURL: "https://bitspark.de/slang/docs/operator/uuid",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "trigger",
				},
				Out: core.TypeDef{
					Type:    "string",
				},
			},
		},
		PropertyDefs: map[string]*core.TypeDef{},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			if i := in.Pull(); !core.IsMarker(i) {
				out.Push(uuid.New().String())
			} else {
				out.Push(i)
			}
		}
	},
}
