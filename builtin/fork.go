package builtin

import (
	"slang/op"
)

func createOpFork(def op.InstanceDef, par *op.Operator) (*op.Operator, error) {
	inDef := op.PortDef{
		Type: "stream",
		Stream: &op.PortDef{
			Type: "map",
			Map: map[string]op.PortDef{
				"i":      {Type: "any"},
				"select": {Type: "boolean"},
			},
		},
	}

	outDef := op.PortDef{
		Type: "map",
		Map: map[string]op.PortDef{
			"true": {
				Type:   "stream",
				Stream: &op.PortDef{Type: "any"},
			},
			"false": {
				Type:   "stream",
				Stream: &op.PortDef{Type: "any"},
			},
		},
	}

	return op.MakeOperator(def.Name, func(in, out *op.Port, store interface{}) {
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
	}, inDef, outDef, par)
}
