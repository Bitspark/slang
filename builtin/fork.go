package builtin

import (
	"slang/op"
	"errors"
)

func createOpFork(def op.InstanceDef, par *op.Operator) (*op.Operator, error) {
	var inDef, outDef op.PortDef

	if def.In == nil || def.Out == nil {
		inDef = op.PortDef{
			Type: "stream",
			Stream: &op.PortDef{
				Type: "map",
				Map: map[string]op.PortDef{
					"i":      {Type: "any"},
					"select": {Type: "boolean"},
				},
			},
		}

		outDef = op.PortDef{
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
	} else {
		if !def.In.Stream.Map["i"].Equals(*def.Out.Map["true"].Stream) {
			return nil, errors.New("in item and true output not equal")
		}
		if !def.In.Stream.Map["i"].Equals(*def.Out.Map["false"].Stream) {
			return nil, errors.New("in item and false output not equal")
		}
		inDef = *def.In
		outDef = *def.Out
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
