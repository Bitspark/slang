package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
)

var syncMergeOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		Services: map[string]*core.ServiceDef{
			core.DEFAULT_SERVICE: {
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
		},
	},
	oFunc: func(srvs map[string]*core.Service, dels map[string]*core.Delegate, store interface{}) {
		in := srvs[core.DEFAULT_SERVICE].In()
		out := srvs[core.DEFAULT_SERVICE].Out()
		for true {
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
