package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

var encodingJSONPathId = uuid.MustParse("89571f57-4aad-4bb8-9d03-573343ff1202")
var encodingJSONPathCfg = &builtinConfig{
	blueprint: core.Blueprint{
		Id: encodingJSONPathId,
		Meta: core.BlueprintMetaDef{
			Name:             "JSON Path",
			ShortDescription: "select values based json path expression from a JSON document",
			Icon:             "brackets-curly",
			Tags:             []string{"json", "encoding"},
			DocURL:           "https://bitspark.de/slang/docs/operator/json-path",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "binary",
				},
				Out: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"valid": {
							Type: "boolean",
						},
						"{paths}": {
							Type: "primitive",
						},
					},
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
		PropertyDefs: map[string]*core.TypeDef{
			"paths": {
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
		paths := op.Property("paths").([]interface{})
		for !op.CheckStop() {
			in := in.Pull()
			valid := true
			if core.IsMarker(in) {
				out.Push(in)
				continue
			}

			jsonDoc := []byte(in.(core.Binary))
			if !gjson.ValidBytes(jsonDoc) {
				for _, v := range paths {
					out.Map(v.(string)).Push(nil)
				}
				valid = false
			} else {
				for _, v := range paths {
					res := gjson.GetBytes(jsonDoc, v.(string))
					if !res.Exists() {
						out.Map(v.(string)).Push(nil)
					}
					if res.IsArray() || res.IsObject() {
						out.Map(v.(string)).Push(res.Raw)
					}
					out.Map(v.(string)).Push(res.Value())
				}
			}
			out.Map("valid").Push(valid)
		}
	},
}
