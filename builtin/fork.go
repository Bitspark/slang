package builtin

import "slang/core"

func createOpFork(def core.InstanceDef, par *core.Operator) (*core.Operator, error) {
	inDef := core.PortDef{
		Type: "stream",
		Stream: &core.PortDef{
			Type: "map",
			Map: map[string]core.PortDef{
				"i":      core.PortDef{Type: "any"},
				"select": core.PortDef{Type: "boolean"},
			},
		},
	}

	outDef := core.PortDef{
		Type: "map",
		Map: map[string]core.PortDef{
			"true": core.PortDef{
				Type:   "stream",
				Stream: &core.PortDef{Type: "any"},
			},
			"false": core.PortDef{
				Type:   "stream",
				Stream: &core.PortDef{Type: "any"},
			},
		},
	}

	return core.NewOperator(def.Name, func(in, out *core.Port, store interface{}) {
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
