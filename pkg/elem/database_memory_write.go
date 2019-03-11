package elem

import (
	"github.com/Bitspark/slang/pkg/core"
)

var databaseMemoryWriteCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"key": {
							Type: "string",
						},
						"value": {
							Type:    "generic",
							Generic: "valueType",
						},
					},
				},
				Out: core.TypeDef{
					Type: "trigger",
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
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
		ms := getMemoryStore(store)

		for {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			pair := i.(map[string]interface{})

			ms.mutex.Lock()
			ms.items[pair["key"].(string)] = pair["value"]
			ms.mutex.Unlock()

			out.Push(nil)
		}
	},
}
