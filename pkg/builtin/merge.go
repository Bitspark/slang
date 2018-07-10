package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
)

var mergeOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
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
						"select": {
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
	oFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			i := in.Map("select").Stream().Pull()
			pTrue := in.Map("true").Stream().Pull()
			pFalse := in.Map("false").Stream().Pull()

			if !in.Map("select").OwnBOS(i) {
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
				panic("port select received BOS too early")
			}

			for {
				i := in.Map("select").Stream().Pull()

				if in.Map("select").OwnEOS(i) {
					pTrue := in.Map("true").Stream().Pull()
					pFalse := in.Map("false").Stream().Pull()

					if in.Map("true").OwnEOS(pTrue) && in.Map("false").OwnEOS(pFalse) {
						out.PushEOS()
					} else {
						panic("port select received EOS too early")
					}

					break
				}

				if pSelect, ok := i.(bool); ok {
					var pName string
					if pSelect {
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
	oConnFunc: func(op *core.Operator, dst, src *core.Port) error {
		if dst == op.Main().In().Map("select") {
			op.Main().Out().SetStreamSource(src.StreamSource())
		}
		return nil
	},
}
