package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
)

var forkOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		Services: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "map",
						Map: map[string]*core.TypeDef{
							"item": {
								Type:    "generic",
								Generic: "itemType",
							},
							"select": {
								Type: "boolean",
							},
						},
					},
				},
				Out: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"true": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type:    "generic",
								Generic: "itemType",
							},
						},
						"false": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type:    "generic",
								Generic: "itemType",
							},
						},
					},
				},
			},
		},
	},
	oFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for {
			i := in.Stream().Pull()

			if !in.OwnBOS(i) {
				out.Push(i)
				continue
			}

			out.Map("true").PushBOS()
			out.Map("false").PushBOS()

			for {
				i := in.Stream().Pull()

				if in.OwnEOS(i) {
					out.Map("true").PushEOS()
					out.Map("false").PushEOS()
					break
				}

				if m, ok := i.(map[string]interface{}); ok {
					pI := m["item"]

					pSelect := m["select"].(bool)

					if pSelect {
						out.Map("true").Push(pI)
					} else {
						out.Map("false").Push(pI)
					}
				} else {
					panic("invalid item")
				}
			}
		}
	},
}
