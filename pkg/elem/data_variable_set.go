package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"

	"sync"
)

type variableStore struct {
	mutex *sync.Mutex
	value interface{}
}

var variableStores map[string]*variableStore
var variableMutex *sync.Mutex

func getVariableStore(store string) *variableStore {
	variableMutex.Lock()
	ws, ok := variableStores[store]
	if !ok {
		ws = &variableStore{}
		ws.mutex = &sync.Mutex{}
		ws.value = nil
		variableStores[store] = ws
	}
	variableMutex.Unlock()
	return ws
}

var dataVariableSetId = uuid.MustParse("3be41b5b-5a43-4f94-a7ae-7f0bacc4ae77")
var dataVariableSetCfg = &builtinConfig{
	blueprint: core.Blueprint{
		Id: dataVariableSetId,
		Meta: core.BlueprintMetaDef{
			Name:             "set value",
			ShortDescription: "stores a value for later use",
			Icon:             "inbox-in",
			Tags:             []string{"data"},
			DocURL:           "https://bitspark.de/slang/docs/operator/set-value",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type:    "generic",
					Generic: "valueType",
				},
				Out: core.TypeDef{
					Type: "trigger",
				},
			},
		},
		PropertyDefs: map[string]*core.TypeDef{
			"valueName": {
				Type: "string",
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()

		// Get store
		store := op.Property("valueName").(string)
		vs := getVariableStore(store)

		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
			}

			vs.mutex.Lock()
			vs.value = i
			vs.mutex.Unlock()

			out.Push(nil)
		}
	},
}
