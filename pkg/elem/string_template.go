package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"strings"
	"fmt"
)

var stringTemplateCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Meta: core.OperatorMetaDef{
			Name: "template",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"content": {
							Type: "string",
						},
						"{variables}": {
							Type: "primitive",
						},
					},
				},
				Out: core.TypeDef{
					Type: "string",
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
		PropertyDefs: map[string]*core.TypeDef{
			"variables": {
				Type: "stream",
				Stream: &core.TypeDef{
					Type: "string",
				},
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		vars := op.Property("variables").([]interface{})
		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			data := i.(map[string]interface{})
			content := data["content"].(string)
			for _, v := range vars {
				val := data[v.(string)]
				valStr := fmt.Sprintf("%v", val)
				content = strings.Replace(content, "{" + v.(string) + "}", valStr, -1)
			}

			out.Push(content)
		}
	},
}
