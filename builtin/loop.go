package builtin

import (
	"slang/core"
	"errors"
)

func createOpLoop(def core.InstanceDef, par *core.Operator) (*core.Operator, error) {
	var inDef, outDef core.PortDef

	if def.In == nil || def.Out == nil {
		return nil, errors.New("need port definitions")
	} else {
		iType := def.In.Map["init"]
		if def.In.Map["iteration"].Stream.Map["state"].Equals(iType) != nil {
			return nil, errors.New("in item and true output not equal")
		}
		if def.Out.Map["end"].Equals(iType) != nil {
			return nil, errors.New("in item and true output not equal")
		}
		if def.Out.Map["state"].Stream.Equals(iType) != nil {
			return nil, errors.New("in item and true output not equal")
		}
		inDef = *def.In
		outDef = *def.Out
	}

	return core.NewOperator(def.Name, func(in, out *core.Port, store interface{}) {
		for true {
			i := in.Map("init").Pull()

			// Redirect all markers
			if isMarker(i) {
				out.Map("end").Push(i)
				continue
			}

			out.Map("state").PushBOS()
			out.Map("state").Stream().Push(i)

			oldState := i

			i = in.Map("iteration").Stream().Pull()

			if !in.Map("iteration").OwnBOS(i) {
				panic("expected own BOS")
			}

			for true {
				iter := in.Map("iteration").Stream().Pull().(map[string]interface{})
				newState := iter["state"]
				cont := iter["continue"].(bool)

				if cont {
					out.Map("state").Push(newState)
				} else {
					out.Map("state").PushEOS()
					i = in.Map("iteration").Stream().Pull()
					if !in.Map("iteration").OwnEOS(i) {
						panic("expected own BOS")
					}
					out.Map("end").Push(oldState)
					break
				}

				oldState = newState
			}
		}
	}, inDef, outDef, par)
}
