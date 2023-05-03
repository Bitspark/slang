package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var streamWindow2Id = uuid.MustParse("5b704038-9617-454a-b7a1-2091277cff70")
var streamWindow2Cfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: streamWindow2Id,
		Meta: core.BlueprintMetaDef{
			Name:             "window 2",
			ShortDescription: "",
			Icon:             "window-restore",
			Tags:             []string{"data", "window"},
			DocURL:           "https://bitspark.de/slang/docs/operator/window",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type:    "generic",
					Generic: "itemType",
				},
				Out: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type:    "generic",
						Generic: "itemType",
					},
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
		PropertyDefs: core.PropertyMap{
			"size": {
				Type: "number",
			},
			"stride": {
				Type: "number",
			},
			"fill": {
				Type: "boolean",
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()

		size := int(op.Property("size").(float64))
		stride := int(op.Property("stride").(float64))
		fill := op.Property("fill").(bool)

		items := []interface{}{}
		ignore := 0
		started := fill

		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			ignore--
			if ignore < 0 {
				items = append(items, i)
			}

			if len(items) == size {
				started = true
				out.Push(items)
				/*
				out.PushBOS()
				for _, item := range items {
					out.Stream().Push(item)
				}
				out.PushEOS()
				*/
				ignore = stride - size
				if ignore <= 0 {
					items = items[stride:]
				} else {
					items = []interface{}{}
				}
			} else if !started && len(items)%stride == 0 {
				out.Push(items)

				/*
				out.PushBOS()
				for _, item := range items {
					out.Stream().Push(item)
				}
				out.PushEOS()
				*/
			}
		}
	},
}
