package elem

import (
	"github.com/Bitspark/slang/pkg/core"
)

var streamMapToStreamCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: "d099a1cd-69eb-43a2-b95b-239612c457fc",
		Meta: core.OperatorMetaDef{
			Name: "map to stream",
			ShortDescription: "takes a map and emits a stream of key-value pairs",
			Icon: "cubes",
			Tags: []string{"stream", "convert"},
			DocURL: "https://bitspark.de/slang/docs/operator/map-to-stream",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"{entries}": {
							Type:    "generic",
							Generic: "valueType",
						},
					},
				},
				Out: core.TypeDef{
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
		DelegateDefs: map[string]*core.DelegateDef{},
		PropertyDefs: core.TypeDefMap{
			"entries": {
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
		entries := []string{}
		for _, entry := range op.Property("entries").([]interface{}) {
			entries = append(entries, entry.(string))
		}
		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			im := i.(map[string]interface{})

			out.PushBOS()
			for _, entry := range entries {
				value := im[entry]
				valueMap := make(map[string]interface{})
				valueMap["key"] = entry
				valueMap["value"] = value
				out.Stream().Push(valueMap)
			}
			out.PushEOS()
		}
	},
}
