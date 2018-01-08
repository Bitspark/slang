package builtin

import (
	"slang/core"
)

var loopOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		In: core.PortDef{
			Type: "map",
			Map: map[string]core.PortDef{
				"init": {
					Type: "any",
					Any:  "stateType",
				},
				"iteration": {
					Type: "stream",
					Stream: &core.PortDef{
						Type: "map",
						Map: map[string]core.PortDef{
							"continue": {
								Type: "boolean",
							},
							"state": {
								Type: "any",
								Any:  "stateType",
							},
						},
					},
				},
			},
		},
		Out: core.PortDef{
			Type: "map",
			Map: map[string]core.PortDef{
				"end": {
					Type: "any",
					Any:  "stateType",
				},
				"state": {
					Type: "stream",
					Stream: &core.PortDef{
						Type: "any",
						Any:  "stateType",
					},
				},
			},
		},
	},
	oFunc: func(in, out *core.Port, store interface{}) {
		for true {
			i := in.Map("init").Pull()

			// Redirect all markers
			if isMarker(i) {
				out.Map("end").Push(i)
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
				iter := in.Map("iteration").Stream().Pull().(map[string]interface{})
				newState := iter["state"]
				cont := iter["continue"].(bool)

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
