package builtin

import (
	"slang/core"
)

var syncForkOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		In: core.PortDef{
			Type: "map",
			Map: map[string]*core.PortDef{
				"i": {
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
	oFunc: func(in, out *core.Port, store interface{}) {
		for true {
			item, ok := in.Pull().(map[string]interface{})
			if !ok {
				out.Push(item)
				continue
			}

			if item["select"].(bool) {
				out.Map("true").Push(item["i"])
				out.Map("false").Push(nil)
			} else {
				out.Map("true").Push(nil)
				out.Map("false").Push(item["i"])
			}
		}
	},
}
