package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var streamTransformCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: uuid.MustParse("dce082cb-7272-4e85-b4fa-740778e8ba8d"),
		Meta: core.BlueprintMetaDef{
			Name:             "transform stream",
			ShortDescription: "transforms a stream by iterating it using an iterator delegate",
			Icon:             "code-commit",
			Tags:             []string{"stream"},
			DocURL:           "https://bitspark.de/slang/docs/operator/map-to-stream",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"initial": {
							Type:    "generic",
							Generic: "stateType",
						},
						"items": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type:    "generic",
								Generic: "inItemType",
							},
						},
					},
				},
				Out: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"result": {
							Type:    "generic",
							Generic: "stateType",
						},
						"items": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type:    "generic",
								Generic: "outItemType",
							},
						},
					},
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{
			"iterator": {
				Out: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"state": {
							Type:    "generic",
							Generic: "stateType",
						},
						"item": {
							Type:    "generic",
							Generic: "inItemType",
						},
					},
				},
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"state": {
							Type:    "generic",
							Generic: "stateType",
						},
						"item": {
							Type:    "generic",
							Generic: "outItemType",
						},
					},
				},
			},
		},
		PropertyDefs: map[string]*core.TypeDef{},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()

		iterator := op.Delegate("iterator")
		iteratorOut := iterator.Out()
		iteratorIn := iterator.In()

		for !op.CheckStop() {
			state := in.Map("initial").Pull()
			if core.IsMarker(state) {
				in.Map("items").Pull()
				out.Push(state)
			}

			in.Map("items").PullBOS()
			out.Map("items").PushBOS()

			for {
				item := in.Map("items").Stream().Pull()
				if in.Map("items").OwnEOS(item) {
					break
				}

				iteratorOut.Map("item").Push(item)
				iteratorOut.Map("state").Push(state)

				state = iteratorIn.Map("state").Pull()
				outItem := iteratorIn.Map("item").Pull()

				out.Map("items").Stream().Push(outItem)
			}

			out.Map("items").PushEOS()
			out.Map("result").Push(state)
		}
	},
}
