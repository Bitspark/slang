package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"sync"
)

type windowStore struct {
	mutex *sync.Mutex
	items []interface{}
}

var windowStores map[string]*windowStore
var windowMutex *sync.Mutex

func getWindowStore(store string) *windowStore {
	windowMutex.Lock()
	ws, ok := windowStores[store]
	if !ok {
		ws = &windowStore{}
		ws.mutex = &sync.Mutex{}
		ws.items = []interface{}{}
		windowStores[store] = ws
	}
	windowMutex.Unlock()
	return ws
}

var streamWindowCollectCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type:    "generic",
					Generic: "itemType",
				},
				Out: core.TypeDef{
					Type: "trigger",
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{
		},
		PropertyDefs: map[string]*core.TypeDef{
			"store": {
				Type: "string",
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()

		// Get store
		store := op.Property("store").(string)
		ws := getWindowStore(store)

		for !op.CheckStop() {
			item := in.Pull()
			if core.IsMarker(item) {
				out.Push(item)
				continue
			}

			ws.mutex.Lock()
			ws.items = append(ws.items, item)
			ws.mutex.Unlock()
		}
	},
}
