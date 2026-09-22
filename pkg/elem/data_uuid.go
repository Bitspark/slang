package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var dataUUIDId = uuid.MustParse("a83bf9b2-cf1b-4b14-94c2-ea04d5cf70c0")
var dataUUIDCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: dataUUIDId,
		Meta: core.BlueprintMetaDef{
			Name:             "generate UUID",
			ShortDescription: "generates a random UUID",
			Icon:             "barcode-alt",
			Tags:             []string{"data"},
			DocURL:           "https://bitspark.de/slang/docs/operator/uuid",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "trigger",
				},
				Out: core.TypeDef{
					Type: "string",
				},
			},
		},
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
