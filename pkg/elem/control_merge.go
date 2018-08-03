package elem

import (
	"github.com/Bitspark/slang/pkg/core"
)

var constrolMergeCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
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
				Out: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type:    "generic",
						Generic: "itemType",
					},
				},
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			i := in.Map("control").Stream().Pull()
			pTrue := in.Map("true").Stream().Pull()
			pFalse := in.Map("false").Stream().Pull()

			if !in.Map("control").OwnBOS(i) {
				if pTrue == pFalse == i {
					out.Push(i)
				} else {
					panic("invalid item: expected same BOS")
				}
				continue
			}

			if in.Map("true").OwnBOS(pTrue) && in.Map("false").OwnBOS(pFalse) {
				out.PushBOS()
			} else {
				panic("port control received BOS too early")
			}

			for {
				i := in.Map("control").Stream().Pull()

				if in.Map("control").OwnEOS(i) {
					pTrue := in.Map("true").Stream().Pull()
					pFalse := in.Map("false").Stream().Pull()

					if in.Map("true").OwnEOS(pTrue) && in.Map("false").OwnEOS(pFalse) {
						out.PushEOS()
					} else {
						panic("port control received EOS too early")
					}

					break
				}

				if pcontrol, ok := i.(bool); ok {
					var pName string
					if pcontrol {
						pName = "true"
					} else {
						pName = "false"
					}
					pI := in.Map(pName).Stream().Pull()
					out.Stream().Push(pI)
				} else {
					// Happens when i == OwnEOS --> should never happen
					panic("invalid item 3")
				}
			}

		}
	},
	opConnFunc: func(op *core.Operator, dst, src *core.Port) error {
		if dst == op.Main().In().Map("control") {
			op.Main().Out().SetStreamSource(src.StreamSource())
		}
		return nil
	},
}
