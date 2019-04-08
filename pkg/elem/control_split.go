package elem

import (
	"github.com/Bitspark/slang/pkg/core"
)

var controlSplitId = "fed72b41-2584-424c-8213-1978410ccab6"
var controlSplitCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: controlSplitId,
		Meta: core.OperatorMetaDef{
			Name:             "split",
			ShortDescription: "emits items of a stream to either outgoing stream depending on the control value",
			Icon:             "code-merge",
			Tags:             []string{"control"},
			DocURL:           "https://bitspark.de/slang/docs/operator/split",
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
						"true": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type:    "generic",
								Generic: "itemType",
							},
						},
						"false": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type:    "generic",
								Generic: "itemType",
							},
						},
						"control": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type: "boolean",
							},
						},
					},
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{
			"controller": {
				In: core.TypeDef{
					Type: "boolean",
				},
				Out: core.TypeDef{
					Type:    "generic",
					Generic: "itemType",
				},
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		dlg := op.Delegate("controller")
		dlgOut := dlg.Out()
		dlgIn := dlg.In()

		for !op.CheckStop() {
			i := in.Stream().Pull()
			if !in.OwnBOS(i) {
				out.Push(i)
				continue
			}

			out.Map("true").PushBOS()
			out.Map("false").PushBOS()
			out.Map("control").PushBOS()

			for {
				i = in.Stream().Pull()
				if in.OwnEOS(i) {
					break
				}

				dlgOut.Push(i)
				c := dlgIn.Pull()

				cb, ok := c.(bool)
				if !ok {
					panic("expected bool")
				}

				out.Map("control").Stream().Push(cb)

				if cb {
					out.Map("true").Stream().Push(i)
				} else {
					out.Map("false").Stream().Push(i)
				}
			}

			out.Map("true").PushEOS()
			out.Map("false").PushEOS()
			out.Map("control").PushEOS()
		}
	},
	opConnFunc: func(op *core.Operator, dst, src *core.Port) error {
		if dst == op.Main().In() {
			op.Main().Out().Map("control").SetStreamSource(src.StreamSource())
		}
		return nil
	},
}
