package elem

import (
	"github.com/Bitspark/slang/pkg/core"
)

var constrolSingleMergeCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"true": {
							Type:    "generic",
							Generic: "itemType",
						},
						"false": {
							Type:    "generic",
							Generic: "itemType",
						},
						"select": {
							Type: "boolean",
						},
					},
				},
				Out: core.TypeDef{
					Type:    "generic",
					Generic: "itemType",
				},
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			item := in.Pull()
			m, ok := item.(map[string]interface{})
			if !ok {
				out.Push(item)
				continue
			}

			if m["select"].(bool) {
				out.Push(m["true"])
			} else {
				out.Push(m["false"])
			}
		}
	},
}
