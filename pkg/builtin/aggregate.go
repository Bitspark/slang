package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
	"fmt"
)

var aggregateOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"init": {
							Type:    "generic",
							Generic: "stateType",
						},
						"items": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type:    "generic",
								Generic: "itemType",
							},
						},
					},
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
						Type:    "generic",
						Generic: "stateType",
					},
				},
				Out: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "map",
						Map: map[string]*core.TypeDef{
							"item": {
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
	oFunc: func(op *core.Operator) {
		iIn := op.Delegate("iteration").In()
		iOut := op.Delegate("iteration").Out()
		in := op.Main().In()
		out := op.Main().Out()
		for {
			state := in.Map("init").Pull()

			// Redirect all markers
			if core.IsMarker(state) {
				if !core.IsMarker(in.Map("items").Stream().Pull()) {
					panic("should be marker")
				}
				out.Push(state)

				iOut.Push(state)
				iIn.Pull()

				continue
			}

			in.Map("items").PullBOS()

			iOut.PushBOS()
			iIn.PullBOS()

			for {
				item := in.Map("items").Stream().Pull()

				if core.IsMarker(item) {
					if in.Map("items").OwnEOS(item) {
						iOut.PushEOS()
						iIn.PullEOS()

						out.Push(state)
						break
					} else {
						panic(fmt.Sprintf(op.Name() + ": unknown marker: %#v", item))
					}
				}

				iOut.Stream().Map("item").Push(item)
				iOut.Stream().Map("state").Push(state)

				state = iIn.Stream().Pull()
			}
		}
	},
	oConnFunc: func(op *core.Operator, dst, src *core.Port) error {
		if dst == op.Main().In().Map("items") {
			iOut := op.Delegate("iteration").Out()
			iOut.SetStreamSource(src.StreamSource())
		}
		return nil
	},
}
