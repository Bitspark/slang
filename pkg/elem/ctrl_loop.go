package elem

import (
	"github.com/Bitspark/slang/pkg/core"
)

var loopOpCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type:    "generic",
					Generic: "stateType",
				},
				Out: core.TypeDef{
					Type:    "generic",
					Generic: "stateType",
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{
			"iteration": {
				In: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "map",
						Map: map[string]*core.TypeDef{
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
				Out: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type:    "generic",
						Generic: "stateType",
					},
				},
			},
		},
	},
	opFunc: func(op *core.Operator) {
		iIn := op.Delegate("iteration").In()
		iOut := op.Delegate("iteration").Out()
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
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

			for {
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
