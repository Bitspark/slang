package elem

import (
	"fmt"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var controlSwitchId = uuid.MustParse("cd6fc5c8-5b64-4b1a-9885-59ede141b398")
var controlSwitchCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: controlSwitchId,
		Meta: core.BlueprintMetaDef{
			Name:             "switch",
			ShortDescription: "emits a constant value for each item",
			Icon:             "code-merge",
			Tags:             []string{"stream", "control"},
			DocURL:           "https://bitspark.de/slang/docs/operator/switch",
		},
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
		PropertyDefs: core.PropertyMap{
			"cases": {
				core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type:    "generic",
						Generic: "selectType",
					},
				},
				nil,
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
