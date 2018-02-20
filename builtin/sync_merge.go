package builtin

import (
	"slang/core"
)

var syncMergeOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		In: core.PortDef{
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
				"select": {
					Type: "boolean",
				},
			},
		},
		Out: core.PortDef{
			Type:    "generic",
			Generic: "itemType",
		},
	},
	oFunc: func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
		for true {
			item, ok := in.Pull().(map[string]interface{})
			if !ok {
				out.Push(item)
				continue
			}

			if item["select"].(bool) {
				out.Push(item["true"])
			} else {
				out.Push(item["false"])
			}
		}
	},
}
