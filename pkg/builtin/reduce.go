package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
)

var reduceOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		Services: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.PortDef{
					Type: "stream",
					Stream: &core.PortDef{
						Type:    "generic",
						Generic: "itemType",
					},
				},
				Out: core.PortDef{
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
	oFunc: func(srvs map[string]*core.Service, dels map[string]*core.Delegate, store interface{}) {
		in := srvs[core.MAIN_SERVICE].In()
		out := srvs[core.MAIN_SERVICE].Out()
		sIn := dels["selection"].In()
		sOut := dels["selection"].Out()
		nullValue := store.(valueStore).value
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
	oPropFunc: func(o *core.Operator, props map[string]interface{}) error {
		if v, ok := props["emptyValue"]; ok {
			o.SetStore(valueStore{v})
		} else {
			o.SetStore(valueStore{nil})
		}
		return nil
	},
}
