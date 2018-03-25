package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
)

var plotOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		Services: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.PortDef{
				},
				Out: core.PortDef{
				},
			},
		},
		Delegates: map[string]*core.DelegateDef{},
	},
	oFunc: func(srvs map[string]*core.Service, dels map[string]*core.Delegate, store interface{}) {
		in := srvs[core.MAIN_SERVICE].In()
		out := srvs[core.MAIN_SERVICE].Out()
		for true {
			// TODO: Implement
			out.Push(in.Pull())
		}
	},
}
