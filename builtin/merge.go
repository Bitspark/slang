package builtin

import (
	"errors"
	"slang/core"
)

func createOpMerge(def core.InstanceDef, par *core.Operator) (*core.Operator, error) {
	var inDef, outDef core.PortDef

	if def.In == nil && def.Out == nil {
		inDef = core.PortDef{
			Type: "map",
			Map: map[string]core.PortDef{
				"true": {
					Type:   "stream",
					Stream: &core.PortDef{Type: "any"},
				},
				"false": {
					Type:   "stream",
					Stream: &core.PortDef{Type: "any"},
				},
				"select": {
					Type:   "stream",
					Stream: &core.PortDef{Type: "boolean"},
				},
			},
		}
		outDef = core.PortDef{
			Type:   "stream",
			Stream: &core.PortDef{Type: "any"},
		}

	} else {
		if def.In == nil || def.Out == nil {
			return nil, errors.New("ports In and Out must be either defined or undefined")
		}

		if !def.In.Map["true"].Equals(*def.Out) {
			return nil, errors.New("out item and true output not equal")
		}
		if !def.In.Map["false"].Equals(*def.Out) {
			return nil, errors.New("out item and false output not equal")
		}
		if !def.In.Map["select"].Equals(core.PortDef{Type: "stream", Stream: &core.PortDef{Type: "boolean"}}) {
			return nil, errors.New("select output def not correct")
		}

		inDef = *def.In
		outDef = *def.Out
	}

	return core.NewOperator(def.Name, func(in, out *core.Port, store interface{}) {
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
	}, inDef, outDef, par)
}
