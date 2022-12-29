package elem

import (
	"encoding/json"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var encodingJSONWriteId = uuid.MustParse("d4aabe2d-dee7-409f-b2bb-713ebc836672")
var encodingJSONWriteCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: encodingJSONWriteId,
		Meta: core.BlueprintMetaDef{
			Name:             "encode JSON",
			ShortDescription: "encodes Slang data into a JSON string",
			Icon:             "brackets-curly",
			Tags:             []string{"encoding"},
			DocURL:           "https://bitspark.de/slang/docs/operator/encode-json",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type:    "generic",
					Generic: "itemType",
				},
				Out: core.TypeDef{
					Type: "binary",
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
			b, err := json.Marshal(&i)
			if err != nil {
				panic(err)
			}
			out.Push(core.Binary(b))
		}
	},
}
