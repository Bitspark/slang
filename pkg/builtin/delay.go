package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
	"time"
)

var delayOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		Services: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.PortDef{
					Type: "number",
				},
				Out: core.PortDef{
					Type: "trigger",
				},
			},
		},
	},
	oFunc: func(srvs map[string]*core.Service, dels map[string]*core.Delegate, store interface{}) {
		in := srvs[core.MAIN_SERVICE].In()
		out := srvs[core.MAIN_SERVICE].Out()
		for {
			i, err := in.PullInt()
			if err != nil {
				if !core.IsMarker(i) {
					out.Push(i)
					continue
				}
				panic("expected number")
			}

			<-time.After(time.Millisecond * time.Duration(i))
			out.Push(1)
		}
	},
	oPropFunc: func(op *core.Operator, i map[string]interface{}) error {
		return nil
	},
	oConnFunc: func(dest, src *core.Port) error {
		return nil
	},
}
