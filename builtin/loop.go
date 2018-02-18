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
			},
		},
		Out: core.PortDef{
			Type: "map",
			Map: map[string]*core.PortDef{
				"end": {
					Type:    "generic",
					Generic: "stateType",
				},
			},
		},
		Delegates: map[string]*core.DelegateDef{
			"iteration": {
				In: core.PortDef{
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
				Out: core.PortDef{
					Type: "stream",
					Stream: &core.PortDef{
						Type:    "generic",
						Generic: "stateType",
					},
				},
			},
		},
	},
	oFunc: func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
		iIn := dels["iteration"].In()
		iOut := dels["iteration"].Out()
		for true {
			i := in.Map("init").Pull()

			// Redirect all markers
			if core.IsMarker(i) {
				iOut.Push(i)
				iter := iIn.Stream().Pull()

				if i != iter {
					panic("should be same marker")
				}

				out.Map("end").Push(i)

				continue
			}

			iOut.PushBOS()
			iOut.Stream().Push(i)

			oldState := i

			i = iIn.Stream().Pull()

			if !iIn.OwnBOS(i) {
				panic("expected own BOS")
			}

			for true {
				iter := iIn.Stream().Pull().(map[string]interface{})
				newState := iter["state"]
				cont := iter["continue"].(bool)

				if cont {
					iOut.Push(newState)
				} else {
					iOut.PushEOS()
					i = iIn.Stream().Pull()
					if !iIn.OwnEOS(i) {
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
