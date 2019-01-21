package elem

import (
	"github.com/Bitspark/slang/pkg/core"
)

var streamWindowCfg = &builtinConfig{
	opDef: core.OperatorDef{
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
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()

		size := int(op.Property("size").(float64))
		stride := int(op.Property("stride").(float64))

		for !op.CheckStop() {
			i := in.Stream().Pull()
			if !in.OwnBOS(i) {
				out.Push(i)
				continue
			}

			items := []interface{}{}
			ignore := 0

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
					out.Stream().Push(items)
					ignore = stride - size
					if ignore <= 0 {
						items = items[stride:]
					} else {
						items = []interface{}{}
					}
				}
			}
			out.PushEOS()
		}
	},
}
