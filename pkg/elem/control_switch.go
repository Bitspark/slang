package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"fmt"
)

var controlSwitchCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"item": {
							Type:    "generic",
							Generic: "inType",
						},
						"select": {
							Type:    "generic",
							Generic: "selectType",
						},
					},
				},
				Out: core.TypeDef{
					Type:    "generic",
					Generic: "outType",
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{
			"{cases}": {
				In: core.TypeDef{
					Type:    "generic",
					Generic: "outType",
				},
				Out: core.TypeDef{
					Type:    "generic",
					Generic: "inType",
				},
			},
			"default": {
				In: core.TypeDef{
					Type:    "generic",
					Generic: "outType",
				},
				Out: core.TypeDef{
					Type:    "generic",
					Generic: "inType",
				},
			},
		},
		PropertyDefs: map[string]*core.TypeDef{
			"cases": {
				Type: "stream",
				Stream: &core.TypeDef{
					Type:    "generic",
					Generic: "selectType",
				},
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		cases := make(map[string]*core.Delegate)
		dflt := op.Delegate("default")
		casesProp := op.Property("cases").([]interface{})
		for _, c := range casesProp {
			cs := fmt.Sprintf("%v", c)
			cases[cs] = op.Delegate(cs)
		}
		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			im := i.(map[string]interface{})
			c := fmt.Sprintf("%v", im["select"])
			cs, ok := cases[c]
			if !ok {
				cs = dflt
			}
			cs.Out().Push(im["item"])
			out.Push(cs.In().Pull())
		}
	},
}
