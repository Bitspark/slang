package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var controlSemaphoreVId = uuid.MustParse("dc9b35a3-bd0e-4ca3-99df-4e2689ea5097")
var controlSemaphoreVCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: controlSemaphoreVId,
		Meta: core.BlueprintMetaDef{
			Name:             "semaphore V",
			ShortDescription: "frees a semaphore token",
			Icon:             "traffic-light-go",
			Tags:             []string{"control", "sync"},
			DocURL:           "https://bitspark.de/slang/docs/operator/semaphore-v",
		},
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
		PropertyDefs: core.PropertyMap{
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
