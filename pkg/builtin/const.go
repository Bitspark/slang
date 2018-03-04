package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
	"errors"
	"github.com/Bitspark/slang/pkg/utils"
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
	oFunc: func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
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
			o.SetStore(valueStore{utils.CleanValue(v)})
			return nil
		} else {
			return errors.New("no value given")
		}
	},
}
