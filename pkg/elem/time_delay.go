package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"time"
)

var timeDelayCfg = &builtinConfig{
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
			i, m := in.PullInt()
			if m != nil {
				if core.IsMarker(m) {
					out.Push(m)
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
