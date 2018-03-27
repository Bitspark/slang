package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
	"errors"
	"github.com/Bitspark/slang/pkg/utils"
)

var constOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		Services: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "trigger",
				},
				Out: core.TypeDef{
					Type:    "generic",
					Generic: "valueType",
				},
			},
		},
	},
	oFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		v := op.Property("value")
		for {
			if i := in.Pull(); !core.IsMarker(i) {
				out.Push(v)
			} else {
				out.Push(i)
			}
		}
	},
	oPropFunc: func(props core.Properties) error {
		if v, ok := props["value"]; ok {
			props["value"] = utils.CleanValue(v)
			return nil
		} else {
			return errors.New("no value given")
		}
	},
}
