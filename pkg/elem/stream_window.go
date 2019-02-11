package elem

import (
	"github.com/Bitspark/slang/pkg/core"
)

var streamWindowCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: "5b704038-9617-454a-b7a1-2091277cff69",
		Meta: core.OperatorMetaDef{
			Name: "window",
			ShortDescription: "cuts a stream into windows of a certain size and emits them",
			Icon: "window-restore",
			Tags: []string{"stream", "window"},
			DocURL: "https://bitspark.de/slang/docs/operator/window",
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
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "stream",
						Stream: &core.TypeDef{
							Type:    "generic",
							Generic: "itemType",
						},
					},
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{
		},
		PropertyDefs: map[string]*core.TypeDef{
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

		for !op.CheckStop() {
			i := in.Stream().Pull()
			if !in.OwnBOS(i) {
				out.Push(i)
				continue
			}

			items := []interface{}{}
			ignore := 0
			started := fill

			out.PushBOS()
			for {
				i = in.Stream().Pull()
				if in.OwnEOS(i) {
					break
				}

				ignore--
				if ignore < 0 {
					items = append(items, i)
				}

				if len(items) == size {
					started = true
					out.Stream().Push(items)
					ignore = stride - size
					if ignore <= 0 {
						items = items[stride:]
					} else {
						items = []interface{}{}
					}
				} else if !started && len(items) % stride == 0 {
					out.Stream().Push(items)
				}
			}
			out.PushEOS()
		}
	},
}
