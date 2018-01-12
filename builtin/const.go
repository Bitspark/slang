package builtin

import (
	"slang/core"
	"errors"
)

type constStore struct {
	value interface{}
}

var constOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		In: core.PortDef{
			Type: "empty",
		},
		Out: core.PortDef{
			Type:    "generic",
			Generic: "valueType",
		},},
	oFunc: func(in, out *core.Port, store interface{}) {
		v := store.(constStore).value
		for true {
			out.Push(v)
		}
	},
	oPropFunc: func(o *core.Operator, props map[string]interface{}) error {
		if v, ok := props["value"]; ok {
			o.SetStore(constStore{v})
			return nil
		} else {
			return errors.New("no value given")
		}
	},
}
