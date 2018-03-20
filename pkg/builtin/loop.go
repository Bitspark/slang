package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
)

var loopOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		Services: map[string]*core.ServiceDef{
			core.DEFAULT_SERVICE: {
				In: core.PortDef{
					Type:    "generic",
					Generic: "stateType",
				},
				Out: core.PortDef{
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
	oFunc: func(srvs map[string]*core.Service, dels map[string]*core.Delegate, store interface{}) {
		iIn := dels["iteration"].In()
		iOut := dels["iteration"].Out()
		in := srvs[core.DEFAULT_SERVICE].In()
		out := srvs[core.DEFAULT_SERVICE].Out()
		for true {
			i := in.Pull()

			// Redirect all markers
			if core.IsMarker(i) {
				iOut.Push(i)
				iter := iIn.Stream().Pull()

				if i != iter {
					panic("should be same marker")
				}

				out.Push(i)

				continue
			}

			iOut.PushBOS()
			iOut.Stream().Push(i)

			oldState := i

			iIn.PullBOS()

			for true {
				iter := iIn.Stream().Pull().(map[string]interface{})
				newState := iter["state"]
				cont := iter["continue"].(bool)

				if cont {
					iOut.Push(newState)
				} else {
					iOut.PushEOS()
					iIn.PullEOS()
					out.Push(oldState)
					break
				}

				oldState = newState
			}

		}
	},
}
