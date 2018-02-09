package builtin

import (
	// "errors"
	"slang/core"
	"errors"
)

type valueStore struct {
	value interface{}
}

var constOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		In: core.PortDef{
			Type: "trigger",
		},
		Out: core.PortDef{
			Type:    "generic",
			Generic: "valueType",
		},
	},
	oFunc: func(in, out *core.Port, store interface{}) {
		val := store.(valueStore).value
		for true {
			i := in.Pull()

			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			out.Push(val)
		}
	},
	oPropFunc: func(o *core.Operator, props map[string]interface{}) error {
		value, ok := props["value"]

		if !ok {
			return errors.New("no value given")
		}

		o.SetStore(valueStore{value})

		return nil
	},
}
