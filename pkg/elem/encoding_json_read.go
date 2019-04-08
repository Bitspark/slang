package elem

import (
	"encoding/json"

	"github.com/Bitspark/slang/pkg/core"
)

var encodingJSONReadId = "b79b019f-5efe-4012-9a1d-1f61549ede25"
var encodingJSONReadCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: encodingJSONReadId,
		Meta: core.OperatorMetaDef{
			Name:             "decode JSON",
			ShortDescription: "decodes a JSON string and emits the corresponding Slang data",
			Icon:             "brackets-curly",
			Tags:             []string{"json", "encoding"},
			DocURL:           "https://bitspark.de/slang/docs/operator/decode-json",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "binary",
				},
				Out: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"valid": {
							Type: "boolean",
						},
						"item": {
							Type:    "generic",
							Generic: "itemType",
						},
					},
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		def, _ := op.Define()
		itemDef := def.ServiceDefs[core.MAIN_SERVICE].Out.Map["item"]
		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}
			var obj interface{}
			err := json.Unmarshal([]byte(i.(core.Binary)), &obj)
			if err != nil {
				out.Map("item").Push(nil)
				out.Map("valid").Push(false)
				continue
			}
			obj = core.CleanValue(obj)
			err = itemDef.VerifyData(obj)
			if err == nil {
				out.Map("item").Push(obj)
				out.Map("valid").Push(true)
			} else {
				out.Map("item").Push(nil)
				out.Map("valid").Push(false)
			}
		}
	},
}
