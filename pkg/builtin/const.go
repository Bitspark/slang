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
		Services: map[string]*core.ServiceDef{
			core.DEFAULT_SERVICE: {
				In: core.PortDef{
					Type: "trigger",
				},
				Out: core.PortDef{
					Type:    "generic",
					Generic: "valueType",
				},
			},
		},
	},
	oFunc: func(srvs map[string]*core.Service, dels map[string]*core.Delegate, store interface{}) {
		v := store.(valueStore).value
		in := srvs[core.DEFAULT_SERVICE].In()
		out := srvs[core.DEFAULT_SERVICE].Out()
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
