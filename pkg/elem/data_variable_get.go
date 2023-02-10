package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"

	"time"
)

var dataVariableGetId = uuid.MustParse("b8771c73-cddf-4eb1-a10c-bf78c2552efe")
var dataVariableGetCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: dataVariableGetId,
		Meta: core.BlueprintMetaDef{
			Name:             "get value",
			ShortDescription: "emits a value previously saved for each item",
			Icon:             "inbox-out",
			Tags:             []string{"data"},
			DocURL:           "https://bitspark.de/slang/docs/operator/get-value",
		},
		ServiceDefs: map[string]*core.ServiceDef{
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
		PropertyDefs: core.PropertyMap{
			"valueName": {
				core.TypeDef{
					Type: "string",
				},
				nil,
			},
			"waitForSet": {
				core.TypeDef{
					Type: "boolean",
				},
				nil,
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()

		// Get store
		store := op.Property("valueName").(string)
		wait := op.Property("waitForSet").(bool)

		vs := getVariableStore(store)

		if wait {
			// wait until value has been set
			for {
				vs.mutex.Lock()
				value := vs.value
				vs.mutex.Unlock()

				if value != nil {
					break
				}

				time.Sleep(10 * time.Millisecond)
			}
		}

		for !op.CheckStop() {
			if i := in.Pull(); core.IsMarker(i) {
				out.Push(i)
				continue
			}

			vs.mutex.Lock()
			out.Push(vs.value)
			vs.mutex.Unlock()
		}
	},
}
