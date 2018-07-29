package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"time"
)

var delayOpCfg = &builtinConfig{
	opDef: core.OperatorDef{
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
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			i, err := in.PullInt()
			if err != nil {
				if !core.IsMarker(i) {
					out.Push(i)
					continue
				}
				panic("expected number")
			}

			<-time.After(time.Millisecond * time.Duration(i))
			out.Push(nil)
		}
	},
	opConnFunc: func(op *core.Operator, dst, src *core.Port) error {
		return nil
	},
}
