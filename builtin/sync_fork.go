package builtin

import (
	"slang/core"
)

var syncForkOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		In: core.PortDef{
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
		Out: core.PortDef{
			Type: "map",
			Map: map[string]*core.PortDef{
				"true": {
					Type:    "generic",
					Generic: "itemType",
				},
				"false": {
					Type:    "generic",
					Generic: "itemType",
				},
			},
		},
	},
	oFunc: func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
		for true {
			item := in.Pull()
			m, ok := item.(map[string]interface{})
			if !ok {
				out.Push(item)
				continue
			}

			if m["select"].(bool) {
				out.Map("true").Push(m["item"])
				out.Map("false").Push(nil)
			} else {
				out.Map("true").Push(nil)
				out.Map("false").Push(m["item"])
			}
		}
	},
}
