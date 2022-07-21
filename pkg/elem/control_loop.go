package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var controlLoopId = uuid.MustParse("0b8a1592-1368-44bc-92d5-692acc78b1d3")
var controlLoopCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: controlLoopId,
		Meta: core.BlueprintMetaDef{
			Name:             "loop",
			ShortDescription: "lets an iterator delegate process a state until the controller tells it to stop",
			Icon:             "undo",
			Tags:             []string{"data", "stream"},
			DocURL:           "https://bitspark.de/slang/docs/operator/loop",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type:    "generic",
					Generic: "stateType",
				},
				Out: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"result": {
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
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{
			"iterator": {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"state": {
							Type:    "generic",
							Generic: "stateType",
						},
						"item": {
							Type:    "generic",
							Generic: "itemType",
						},
					},
				},
				Out: core.TypeDef{
					Type:    "generic",
					Generic: "stateType",
				},
			},
			"controller": {
				In: core.TypeDef{
					Type: "boolean",
				},
				Out: core.TypeDef{
					Type:    "generic",
					Generic: "stateType",
				},
			},
		},
	},
	opFunc: func(op *core.Operator) {
		iterIn := op.Delegate("iterator").In()
		iterOut := op.Delegate("iterator").Out()
		ctrlIn := op.Delegate("controller").In()
		ctrlOut := op.Delegate("controller").Out()
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			s := in.Pull()

			// Redirect all markers
			if core.IsMarker(s) {
				out.Push(s)
				continue
			}

			out.Map("items").PushBOS()

			for {
				// Ask controller whether to continue
				ctrlOut.Push(s)
				cont, _ := ctrlIn.PullBoolean()
				if !cont {
					break
				}

				// Let iterator calculate new state
				iterOut.Push(s)

				// Retrieve new state from iterator
				ns := iterIn.Pull().(map[string]interface{})

				// Set state for next iteration
				s = ns["state"]

				// Emit stream item
				out.Map("items").Push(ns["item"])
			}

			out.Map("result").Push(s)
			out.Map("items").PushEOS()
		}
	},
}
