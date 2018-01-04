package builtin

import (
	"fmt"
	"slang/op"
)

func createOperatorFork(def op.InstanceDef, par *op.Operator) (*op.Operator, error) {
	inDef := op.PortDef{
		Type: "map",
		Map: map[string]op.PortDef{
			"i":      op.PortDef{Type: "any"},
			"select": op.PortDef{Type: "boolean"},
		},
	}

	outDef := op.PortDef{
		Type: "map",
		Map: map[string]op.PortDef{
			"true":  op.PortDef{Type: "any"},
			"false": op.PortDef{Type: "any"},
		},
	}

	return op.MakeOperator(def.Name, func(in, out *op.Port, store interface{}) {
		for true {
			i := in.Pull()
			if m, ok := i.(map[string]interface{}); ok {
				fmt.Println(">>>", m)
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
