package builtin

import (
	"slang/core"
)

var reduceOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		In: core.PortDef{
			Type: "map",
			Map: map[string]*core.PortDef{
				"items": {
					Type: "stream",
					Stream: &core.PortDef{
						Type:    "generic",
						Generic: "itemType",
					},
				},
			},
		},
		Out: core.PortDef{
			Type: "map",
			Map: map[string]*core.PortDef{
				"result": {
					Type:    "generic",
					Generic: "itemType",
				},
			},
		},
		Delegates: map[string]*core.DelegateDef{
			"selection": {
				In: core.PortDef{
					Type: "stream",
					Stream: &core.PortDef{
						Type:    "generic",
						Generic: "itemType",
					},
				},
				Out: core.PortDef{
					Type: "stream",
					Stream: &core.PortDef{
						Type: "map",
						Map: map[string]*core.PortDef{
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
	oFunc: func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
		sIn := dels["selection"].In()
		sOut := dels["selection"].Out()
		nullValue := store.(valueStore).value
		for true {
			i := in.Map("items").Pull()

			if core.IsMarker(i) {
				sOut.Push(i)
				sel := sIn.Pull()

				if sel != i {
					panic("expected different marker")
				}

				out.Map("result").Push(i)

				continue
			}

			items, ok := i.([]interface{})
			if !ok {
				panic("expected stream")
			}

			if len(items) == 0 {
				out.Map("result").Push(nullValue)
				continue
			}

			if len(items) == 1 {
				out.Map("result").Push(items[0])
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

			for true {
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
					out.Map("result").Push(items[0])
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
	oPropFunc: func(o *core.Operator, props map[string]interface{}) error {
		if v, ok := props["emptyValue"]; ok {
			o.SetStore(valueStore{v})
		} else {
			o.SetStore(valueStore{nil})
		}
		return nil
	},
}
