package elem

import (
	"fmt"
	"strconv"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var streamParallelizeId = uuid.MustParse("b8428777-7667-4012-b76a-a5b7f4d1e433")
var streamParallelizeCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: streamParallelizeId,
		Meta: core.BlueprintMetaDef{
			Name:             "parallelize",
			ShortDescription: "takes a stream and emits a map of items, selected by given indices",
			Icon:             "align-justify",
			Tags:             []string{"stream"},
			DocURL:           "https://bitspark.de/slang/docs/operator/parallelize",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type:    "generic",
						Generic: "itemType",
					},
				},
				Out: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"el_{indexes}": {
							Type:    "generic",
							Generic: "itemType",
						},
					},
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
		PropertyDefs: core.PropertyMap{
			"indexes": {
				core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "number",
					},
				},
				nil,
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		indexesProp := op.Property("indexes").([]interface{})
		var indexes []int
		for _, idxProp := range indexesProp {
			if idx, ok := idxProp.(int); ok {
				indexes = append(indexes, idx)
			} else if idx, ok := idxProp.(float64); ok {
				indexes = append(indexes, int(idx))
			} else {
				idx, _ := strconv.Atoi(fmt.Sprintf("%v", idx))
				indexes = append(indexes, idx)
			}
		}
		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			array := i.([]interface{})
			arrayLen := len(array)
			for _, idx := range indexes {
				mapEntry := "el_" + strconv.Itoa(idx)
				if idx < arrayLen {
					out.Map(mapEntry).Push(array[idx])
				} else {
					out.Map(mapEntry).Push(nil)
				}
			}
		}
	},
}
