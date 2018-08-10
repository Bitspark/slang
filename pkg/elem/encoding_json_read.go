package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"encoding/json"
	"github.com/Bitspark/slang/pkg/utils"
)

var encodingJSONReadCfg = &builtinConfig{
	opDef: core.OperatorDef{
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
							Type: "generic",
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
			err := json.Unmarshal([]byte(i.(utils.Binary)), &obj)
			if err != nil {
				out.Map("item").Push(nil)
				out.Map("valid").Push(false)
				continue
			}
			obj = utils.CleanValue(obj)
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
