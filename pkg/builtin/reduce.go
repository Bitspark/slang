package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
)

var reduceOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type:    "generic",
						Generic: "itemType",
					},
				},
				Out: core.TypeDef{
					Type:    "generic",
					Generic: "itemType",
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{
			"selection": {
				In: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type:    "generic",
						Generic: "itemType",
					},
				},
				Out: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "map",
						Map: map[string]*core.TypeDef{
							"a": {
								Type:    "generic",
								Generic: "itemType",
							},
							"b": {
								Type:    "generic",
								Generic: "itemType",
							},
						},
					},
				},
			},
		},
	},
	oFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		sIn := op.Delegate("selection").In()
		sOut := op.Delegate("selection").Out()
		nullValue := op.Property("emptyValue")
		for {
			i := in.Pull()

			if core.IsMarker(i) {
				sOut.Push(i)
				sel := sIn.Pull()

				if sel != i {
					panic("expected different marker")
				}

				out.Push(i)

				continue
			}

			items, ok := i.([]interface{})
			if !ok {
				panic("expected stream")
			}

			if len(items) == 0 {
				out.Push(nullValue)
				continue
			}

			if len(items) == 1 {
				out.Push(items[0])
				continue
			}

			sOut.PushBOS()
			j := 0
			for j+1 < len(items) {
				sOut.Stream().Map("a").Push(items[j])
				sOut.Stream().Map("b").Push(items[j+1])
				j += 2
			}
			sOut.PushEOS()

			var leftover interface{}
			if j != len(items) {
				leftover = items[len(items)-1]
			}

			// POOL

			for {
				p := sIn.Pull()

				items, ok := p.([]interface{})
				if !ok {
					panic("expected stream")
				}

				if leftover != nil {
					items = append([]interface{}{leftover}, items...)
				}

				if len(items) == 0 {
					panic("empty pool")
				}

				if len(items) == 1 {
					out.Push(items[0])
					break
				}

				sOut.PushBOS()
				j := 0
				for j+1 < len(items) {
					sOut.Stream().Map("a").Push(items[j])
					sOut.Stream().Map("b").Push(items[j+1])
					j += 2
				}
				sOut.PushEOS()

				if j != len(items) {
					leftover = items[len(items)-1]
				} else {
					leftover = nil
				}
			}
		}
	},
}
