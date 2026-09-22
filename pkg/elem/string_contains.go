package elem

import (
	"strings"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var stringContainsCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: uuid.MustParse("8a01dfe3-5dcf-4f40-9e54-f5b168148d2a"),
		Meta: core.BlueprintMetaDef{
			Name:             "contains",
			ShortDescription: "tells if a string contains another string",
			Icon:             "search",
			Tags:             []string{"string"},
			DocURL:           "https://bitspark.de/slang/docs/operator/contains",
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
			dataOut := strings.Contains(str, subStr)

			out.Push(dataOut)
		}
	},
}
