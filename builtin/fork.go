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
				"i":      op.PortDef{Type: "any"},
				"select": op.PortDef{Type: "boolean"},
			},
		},
	}

	outDef := op.PortDef{
		Type: "map",
		Map: map[string]op.PortDef{
			"true": op.PortDef{
				Type:   "stream",
				Stream: &op.PortDef{Type: "any"},
			},
			"false": op.PortDef{
				Type:   "stream",
				Stream: &op.PortDef{Type: "any"},
			},
		},
	}

	return op.MakeOperator(def.Name, func(in, out *op.Port, store interface{}) {
		for true {
			i := in.Stream().Pull()

			if m, ok := i.(map[string]interface{}); ok {
				pI := m["i"]

				pSelect := m["select"].(bool)

				if pSelect {
					out.Map("true").Push(pI)
				} else {
					out.Map("false").Push(pI)
				}
			} else {
				out.Push(i)
			}
		}
	}, inDef, outDef, par)
}
