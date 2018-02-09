package builtin

import (
	"slang/core"
)

var loopOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		In: core.PortDef{
			Type: "map",
			Map: map[string]*core.PortDef{
				"init": {
					Type:    "generic",
					Generic: "stateType",
				},
				"iteration": {
					Type: "stream",
					Stream: &core.PortDef{
						Type: "map",
						Map: map[string]*core.PortDef{
							"continue": {
								Type: "boolean",
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
		Out: core.PortDef{
			Type: "map",
			Map: map[string]*core.PortDef{
				"end": {
					Type:    "generic",
					Generic: "stateType",
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
	},
	oFunc: func(in, out *core.Port, store interface{}) {
		for true {
			i := in.Map("init").Pull()

			// Redirect all markers
			if core.IsMarker(i) {
				out.Map("end").Push(i)
				out.Map("state").Push(i)
				continue
			}

			out.Map("state").PushBOS()
			out.Map("state").Stream().Push(i)

			oldState := i

			i = in.Map("iteration").Stream().Pull()

			if !in.Map("iteration").OwnBOS(i) {
				panic("expected own BOS")
			}

			for true {
				iter := in.Map("iteration").Stream().Pull()

				if core.IsMarker(iter) {
					continue
				}

				iterMap := iter.(map[string]interface{})
				newState := iterMap["state"]
				cont := iterMap["continue"].(bool)

				if cont {
					out.Map("state").Push(newState)
				} else {
					out.Map("state").PushEOS()
					i = in.Map("iteration").Stream().Pull()
					if !in.Map("iteration").OwnEOS(i) {
						panic("expected own BOS")
					}
					out.Map("end").Push(oldState)
					break
				}

				oldState = newState
			}
		}
	},
}
