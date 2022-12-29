package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

var encodingJSONPathId = uuid.MustParse("89571f57-4aad-4bb8-9d03-573343ff1202")
var encodingJSONPathCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: encodingJSONPathId,
		Meta: core.BlueprintMetaDef{
			Name:             "JSON Path",
			ShortDescription: "select values based json path expression from a JSON document",
			Icon:             "brackets-curly",
			Tags:             []string{"encoding"},
			DocURL:           "https://bitspark.de/slang/docs/operator/json-path",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"document": {
							Type: "binary",
						},
						"{path_names}": {
							Type: "string",
						},
					},
				},
				Out: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"valid": {
							Type: "boolean",
						},
						"{path_names}": {
							Type: "primitive",
						},
					},
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
		PropertyDefs: map[string]*core.TypeDef{
			"path_names": {
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
		path_names := op.Property("path_names").([]interface{})
		for !op.CheckStop() {
			in := in.Pull()
			valid := true
			if core.IsMarker(in) {
				out.Push(in)
				continue
			}

			data := in.(map[string]interface{})
			jsonDoc := []byte(data["document"].(core.Binary))

			if !gjson.ValidBytes(jsonDoc) {
				for _, path_name := range path_names {
					out.Map(path_name.(string)).Push(nil)
				}
				valid = false
			} else {
				for _, path_name := range path_names {
					res := gjson.GetBytes(jsonDoc, data[path_name.(string)].(string))
					if !res.Exists() {
						out.Map(path_name.(string)).Push(nil)
					}
					if res.IsArray() || res.IsObject() {
						out.Map(path_name.(string)).Push(res.Raw)
					}
					out.Map(path_name.(string)).Push(res.Value())
				}
			}
			out.Map("valid").Push(valid)
		}
	},
}
