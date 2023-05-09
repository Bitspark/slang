package elem

import (
	"fmt"

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
			Name:             "JSONPath",
			ShortDescription: "query values from JSON",
			Icon:             "brackets-curly",
			Tags:             []string{"encoding"},
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
						"{path_names.#.name}": {
							Type:    "generic",
							Generic: "itemType",
						},
					},
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
		PropertyDefs: core.PropertyMap{
			"path_names": {
				Type: "stream",
				Stream: &core.TypeDef{
					Type: "map",
					Map: core.TypeDefMap{
						"query": {Type: "string"},
						"name": {Type: "string"},
					},
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

			data := in
			fmt.Println("1)", data)
			jsonDoc := []byte(data.(core.Binary))

			if !gjson.ValidBytes(jsonDoc) {
				for _, path_name := range path_names {
					fmt.Println("   2)", path_name.(map[string]interface{}))
					out.Map(path_name.(map[string]interface{})["name"].(string)).Push(nil)
				}
				valid = false
			} else {
				for _, path_name := range path_names {
					res := gjson.GetBytes(jsonDoc, path_name.(map[string]interface{})["query"].(string))
					if !res.Exists() {
					fmt.Println("   3)", path_name.(map[string]interface{}))
						out.Map(path_name.(map[string]interface{})["name"].(string)).Push(nil)
					}
					/*
					if res.IsArray() || res.IsObject() {
						out.Map(path_name.(map[string]interface{})["name"].(string)).Push(res.Raw)
					}
					*/
					fmt.Println("   4)", path_name.(map[string]interface{}))
					out.Map(path_name.(map[string]interface{})["name"].(string)).Push(res.Value())
				}
			}
			out.Map("valid").Push(valid)
		}
	},
}
