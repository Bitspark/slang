package elem

import (
	"fmt"
	"strings"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var stringTemplateId = uuid.MustParse("3c39f999-b5c2-490d-aed1-19149d228b04")
var stringTemplateCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: stringTemplateId,
		Meta: core.BlueprintMetaDef{
			Name:             "template",
			ShortDescription: "replaces placeholders in a given string with given values",
			Icon:             "stamp",
			Tags:             []string{"string"},
			DocURL:           "https://bitspark.de/slang/docs/operator/template",
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
		PropertyDefs: core.PropertyMap{
			"variables": {
				core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "string",
					},
				},
				nil,
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
				content = strings.Replace(content, "{"+v.(string)+"}", valStr, -1)
			}

			out.Push(content)
		}
	},
}
