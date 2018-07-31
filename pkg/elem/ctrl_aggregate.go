package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"fmt"
)

var aggregateOpCfg = &builtinConfig{
	opDef: core.OperatorDef{
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
			"iterator": {
				In: core.TypeDef{
					Type:    "generic",
					Generic: "stateType",
				},
				Out: core.TypeDef{
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
	opFunc: func(op *core.Operator) {
		iterIn := op.Delegate("iterator").In()
		iterOut := op.Delegate("iterator").Out()
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			state := in.Map("init").Pull()

			// Redirect all markers
			if core.IsMarker(state) {
				if !core.IsMarker(in.Map("items").Stream().Pull()) {
					panic("should be marker")
				}
				out.Push(state)

				iterOut.Push(state)
				iterIn.Pull()

				continue
			}

			in.Map("items").PullBOS()

			for {
				item := in.Map("items").Stream().Pull()

				if core.IsMarker(item) {
					if in.Map("items").OwnEOS(item) {
						out.Push(state)
						break
					} else {
						panic(fmt.Sprintf(op.Name() + ": unknown marker: %#v", item))
					}
				}

				iterOut.Map("item").Push(item)
				iterOut.Map("state").Push(state)

				state = iterIn.Pull()
			}
		}
	},
	opConnFunc: func(op *core.Operator, dst, src *core.Port) error {
		return nil
	},
}
