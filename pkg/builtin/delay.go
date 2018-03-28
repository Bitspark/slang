package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
	"time"
)

var delayOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "number",
				},
				Out: core.TypeDef{
					Type: "trigger",
				},
			},
		},
	},
	oFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
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
	oConnFunc: func(op *core.Operator, dst, src *core.Port) error {
		return nil
	},
}
