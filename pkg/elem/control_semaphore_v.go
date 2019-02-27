package elem

import (
	"github.com/Bitspark/slang/pkg/core"
)

var controlSemaphoreVCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type:    "generic",
					Generic: "itemType",
				},
				Out: core.TypeDef{
					Type:    "generic",
					Generic: "itemType",
				},
			},
		},
		PropertyDefs: map[string]*core.TypeDef{
			"semaphore": {
				Type: "string",
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()

		sem := op.Property("semaphore").(string)
		semStore := getSemaphoreStore(sem)

		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			<-semStore.semaphore
			out.Push(i)
		}
	},
}
