package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
)

var forkOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		In: core.PortDef{
			Type: "stream",
			Stream: &core.PortDef{
				Type: "map",
				Map: map[string]*core.PortDef{
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
		Out: core.PortDef{
			Type: "map",
			Map: map[string]*core.PortDef{
				"true": {
					Type: "stream",
					Stream: &core.PortDef{
						Type:    "generic",
						Generic: "itemType",
					},
				},
				"false": {
					Type: "stream",
					Stream: &core.PortDef{
						Type:    "generic",
						Generic: "itemType",
					},
				},
			},
		},
	},
	oFunc: func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
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
