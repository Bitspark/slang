package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"strconv"
	"fmt"
)

var streamParallelizeCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: "b8428777-7667-4012-b76a-a5b7f4d1e433",
		Meta: core.OperatorMetaDef{
			Name: "parallelize",
			ShortDescription: "takes a stream and emits a map of items, selected by given indices",
			Icon: "align-justify",
			Tags: []string{"stream", "convert"},
			DocURL: "https://bitspark.de/slang/docs/operator/parallelize",
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
		PropertyDefs: core.TypeDefMap{
			"indexes": {
				Type: "stream",
				Stream: &core.TypeDef{
					Type: "number",
				},
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
