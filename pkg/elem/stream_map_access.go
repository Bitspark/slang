package elem

import (
	"reflect"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var streamMapAccessId = uuid.MustParse("618c4007-70fc-44ac-9443-184df77ab730")
var streamMapAccessCfg = &builtinConfig{
	blueprint: core.Blueprint{
		Id: streamMapAccessId,
		Meta: core.BlueprintMetaDef{
			Name:             "access map",
			ShortDescription: "takes a stream of key-value pairs and emits the value with the given key",
			Icon:             "envelope",
			Tags:             []string{"stream"},
			DocURL:           "https://bitspark.de/slang/docs/operator/access-map",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"key": {
							Type:    "generic",
							Generic: "keyType",
						},
						"stream": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type: "map",
								Map: map[string]*core.TypeDef{
									"key": {
										Type:    "generic",
										Generic: "keyType",
									},
									"value": {
										Type:    "generic",
										Generic: "valueType",
									},
								},
							},
						},
					},
				},
				Out: core.TypeDef{
					Type:    "generic",
					Generic: "valueType",
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
		PropertyDefs: core.TypeDefMap{},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
		start:
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			im := i.(map[string]interface{})
			key := core.CleanValue(im["key"])
			stream := im["stream"].([]interface{})

			for _, el := range stream {
				elm := el.(map[string]interface{})
				ckey := core.CleanValue(elm["key"])

				if reflect.DeepEqual(key, ckey) {
					out.Push(elm["value"])
					goto start
				}
			}

			out.Push(nil)
		}
	},
}
