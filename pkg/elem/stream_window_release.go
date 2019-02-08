package elem

import (
	"github.com/Bitspark/slang/pkg/core"
)

var streamWindowReleaseCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: "47b3f097-2043-42c6-aad5-0cfdb9004aef",
		Meta: core.OperatorMetaDef{
			Name: "release window",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "trigger",
				},
				Out: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type:    "generic",
						Generic: "itemType",
					},
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{
		},
		PropertyDefs: map[string]*core.TypeDef{
			"store": {
				Type: "string",
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()

		// Get store
		store := op.Property("store").(string)
		ws := getWindowStore(store)

		for !op.CheckStop() {
			item := in.Pull()
			if core.IsMarker(item) {
				out.Push(item)
				continue
			}

			ws.mutex.Lock()
			out.Push(ws.items)
			ws.items = []interface{}{}
			ws.mutex.Unlock()
		}
	},
}
