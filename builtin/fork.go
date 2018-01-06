package builtin

import (
	"errors"
	"slang/core"
)

func createOpFork(def core.InstanceDef, par *core.Operator) (*core.Operator, error) {
	var inDef, outDef core.PortDef

	if def.In == nil && def.Out == nil {
		inDef = core.PortDef{
			Type: "stream",
			Stream: &core.PortDef{
				Type: "map",
				Map: map[string]core.PortDef{
					"i":      {Type: "any", Any: "itemType"},
					"select": {Type: "boolean"},
				},
			},
		}

		outDef = core.PortDef{
			Type: "map",
			Map: map[string]core.PortDef{
				"true": {
					Type:   "stream",
					Stream: &core.PortDef{Type: "any", Any: "itemType"},
				},
				"false": {
					Type:   "stream",
					Stream: &core.PortDef{Type: "any", Any: "itemType"},
				},
			},
		}
	} else {
		if def.In == nil || def.Out == nil {
			return nil, errors.New("ports In and Out must be either defined or undefined")
		}

		if def.In.Stream.Map["i"].Equals(*def.Out.Map["true"].Stream) != nil {
			return nil, errors.New("in item and true output not equal")
		}
		if def.In.Stream.Map["i"].Equals(*def.Out.Map["false"].Stream) != nil {
			return nil, errors.New("in item and false output not equal")
		}
		inDef = *def.In
		outDef = *def.Out
	}

	return core.NewOperator(def.Name, func(in, out *core.Port, store interface{}) {
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
