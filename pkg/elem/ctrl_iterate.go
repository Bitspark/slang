package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"fmt"
)

var constrolIterateCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"initial": {
							Type:    "generic",
							Generic: "stateType",
						},
						"items": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type:    "generic",
								Generic: "inItemType",
							},
						},
					},
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
								Generic: "outItemType",
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
						"item": {
							Type:    "generic",
							Generic: "outItemType",
						},
						"state": {
							Type:    "generic",
							Generic: "stateType",
						},
					},
				},
				Out: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"item": {
							Type:    "generic",
							Generic: "inItemType",
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
			state := in.Map("initial").Pull()

			// Redirect all markers
			if core.IsMarker(state) {
				if !core.IsMarker(in.Map("items").Stream().Pull()) {
					panic("should be marker")
				}
				out.Push(state)
				continue
			}

			in.Map("items").PullBOS()
			out.Map("items").PushBOS()

			for {
				inItem := in.Map("items").Stream().Pull()
				if core.IsMarker(inItem) {
					if in.Map("items").OwnEOS(inItem) {
						break
					}
					panic(fmt.Sprintf(op.Name()+": unknown marker: %#v", inItem))
				}

				iterOut.Map("item").Push(inItem)
				iterOut.Map("state").Push(state)

				state = iterIn.Map("state").Pull()
				out.Map("items").Push(iterIn.Map("item").Pull())
			}

			out.Map("result").Push(state)
			out.Map("items").PushEOS()
		}
	},
	opConnFunc: func(op *core.Operator, dst, src *core.Port) error {
		if dst == op.Main().In().Map("items") {
			op.Main().Out().Map("items").SetStreamSource(src.StreamSource())
		}
		return nil
	},
}
