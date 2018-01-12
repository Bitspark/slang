package builtin

import (
	"slang/core"
)

var aggOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		In: core.PortDef{
			Type: "map",
			Map: map[string]*core.PortDef{
				"init": {
					Type:    "generic",
					Generic: "stateType",
				},
				"items": {
					Type: "stream",
					Stream: &core.PortDef{
						Type:    "generic",
						Generic: "itemType",
					},
				},
				"state": {
					Type: "stream",
					Stream: &core.PortDef{
						Type:    "generic",
						Generic: "stateType",
					},
				},
			},
		},
		Out: core.PortDef{
			Type: "map",
			Map: map[string]*core.PortDef{
				"end": {
					Type:    "generic",
					Generic: "stateType",
				},
				"iteration": {
					Type: "stream",
					Stream: &core.PortDef{
						Type: "map",
						Map: map[string]*core.PortDef{
							"i": {
								Type:    "generic",
								Generic: "itemType",
							},
							"state": {
								Type:    "generic",
								Generic: "stateType",
							},
						},
					},
				},
			},
		},
	},
	oFunc: func(in, out *core.Port, store interface{}) {
		for true {
			state := in.Map("init").Pull()

			// Redirect all markers
			if core.IsMarker(state) {
				if !core.IsMarker(in.Map("items").Stream().Pull()) {
					panic("should be marker")
				}
				out.Map("end").Push(state)
				continue
			}

			e := in.Map("items").Stream().Pull()
			if !in.Map("items").OwnBOS(e) {
				panic("expected BOS")
			}

			out.Map("iteration").PushBOS()

			if !in.Map("state").OwnBOS(in.Map("state").Stream().Pull()) {
				panic("expected own BOS")
			}

			for true {
				item := in.Map("items").Stream().Pull()

				if core.IsMarker(item) {
					if in.Map("items").OwnEOS(item) {
						out.Map("iteration").PushEOS()
						item = in.Map("state").Stream().Pull()
						if !in.Map("state").OwnEOS(item) {
							panic("expected own BOS")
						}
						out.Map("end").Push(state)
						break
					} else {
						panic("unexpected unknown marker")
					}
				}

				out.Map("iteration").Stream().Map("i").Push(item)
				out.Map("iteration").Stream().Map("state").Push(state)

				state = in.Map("state").Stream().Pull()
			}
		}
	},
}
