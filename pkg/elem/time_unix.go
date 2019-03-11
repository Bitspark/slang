package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"time"
)

var timeUNIXMillisCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "trigger",
				},
				Out: core.TypeDef{
					Type: "number",
				},
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			if i := in.Pull(); !core.IsMarker(i) {
				out.Push(float64(time.Now().UnixNano() / 1000 / 1000))
			} else {
				out.Push(i)
			}
		}
	},
}
