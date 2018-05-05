package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
	"fmt"
)

var switchOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
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
	oFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		cases := make(map[string]*core.Delegate)
		casesProp := op.Property("cases").([]interface{})
		for _, c := range casesProp {
			cs := fmt.Sprintf("%v", c)
			cases[cs] = op.Delegate(cs)
		}
		for {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)

				for _, dlg := range cases {
					dlg.Out().Push(i)
				}
				for _, dlg := range cases {
					dlg.In().Pull()
				}

				continue
			}

			im := i.(map[string]interface{})
			c := fmt.Sprintf("%v", im["select"])
			cases[c].Out().Push(im["item"])
			out.Push(cases[c].In().Pull())
		}
	},
}
