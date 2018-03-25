package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
	"time"
)

var delayOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		In: core.PortDef{
			Type: "number",
		},
		Out: core.PortDef{
			Type: "trigger",
		},
	},
	oFunc: func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
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
