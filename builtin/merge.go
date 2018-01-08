package builtin

import (
	"slang/core"
)

var mergeOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		In: core.PortDef{
			Type: "map",
			Map: map[string]*core.PortDef{
				"true": {
					Type: "stream",
					Stream: &core.PortDef{
						Type:    "generic",
						Generic: "itemType",
					},
				},
				"false": {
					Type: "stream",
					Stream: &core.PortDef{
						Type:    "generic",
						Generic: "itemType",
					},
				},
				"select": {
					Type: "stream",
					Stream: &core.PortDef{
						Type: "boolean",
					},
				},
			},
		},
		Out: core.PortDef{
			Type: "stream",
			Stream: &core.PortDef{
				Type:    "generic",
				Generic: "itemType",
			},
		},
	},
	oFunc: func(in, out *core.Port, store interface{}) {
		for true {
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

			for true {
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
}
