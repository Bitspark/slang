package builtin

import (
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
		v := store.(valueStore).value
		for true {
			if i := in.Pull(); !core.IsMarker(i) {
				out.Push(v)
			} else {
				out.Push(i)
			}
		}
	},
	oPropFunc: func(o *core.Operator, props map[string]interface{}) error {
		if v, ok := props["value"]; ok {
			o.SetStore(valueStore{v})
			return nil
		} else {
			return errors.New("no value given")
		}
	},
}
