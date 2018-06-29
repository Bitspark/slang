package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
	"strings"
	"fmt"
	"github.com/Bitspark/slang/pkg/utils"
)

var templateFormatOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"content": {
							Type: "binary",
						},
						"{variables}": {
							Type: "primitive",
						},
					},
				},
				Out: core.TypeDef{
					Type: "binary",
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
	oFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		vars := op.Property("variables").([]interface{})
		for {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			data := i.(map[string]interface{})
			content := string(data["content"].(utils.Binary))
			for _, v := range vars {
				val := data[v.(string)]
				valStr := fmt.Sprintf("%v", val)
				content = strings.Replace(content, "{" + v.(string) + "}", valStr, -1)
			}

			out.Push(utils.Binary(content))
		}
	},
}
