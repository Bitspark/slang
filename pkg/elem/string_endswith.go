package elem

import (
	"strings"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var stringEndswithCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: uuid.MustParse("db8b1677-baaf-4072-8047-0359cd68be9e"),
		Meta: core.OperatorMetaDef{
			Name:             "ends with",
			ShortDescription: "tells if a string ends with another string",
			Icon:             "hand-point-right",
			Tags:             []string{"string"},
			DocURL:           "https://bitspark.de/slang/docs/operator/ends-with",
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
			dataOut := strings.HasSuffix(str, subStr)

			out.Push(dataOut)
		}
	},
}
