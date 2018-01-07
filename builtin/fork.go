package builtin

import (
	"slang/core"
)

var forkOpCfg = &builtinConfig{
	oFunc: func(in, out *core.Port, store interface{}) {
		for true {
			i := in.Stream().Pull()

			if !in.OwnBOS(i) {
				out.Push(i)
			}

			out.Map("true").PushBOS()
			out.Map("false").PushBOS()

			for true {
				i := in.Stream().Pull()

				if in.OwnEOS(i) {
					out.Map("true").PushEOS()
					out.Map("false").PushEOS()
					break
				}

				if m, ok := i.(map[string]interface{}); ok {
					pI := m["i"]

					pSelect := m["select"].(bool)

					if pSelect {
						out.Map("true").Push(pI)
					} else {
						out.Map("false").Push(pI)
					}
				} else {
					panic("invalid item")
				}
			}
		}
	},
	oDef: nil,
}
