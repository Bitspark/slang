package elem

import (
	"sync"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
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
	safe: true,
	blueprint: core.Blueprint{
		Id: uuid.MustParse("14f5de1a-5e38-4f9c-a625-eff7a572078c"),
		Meta: core.BlueprintMetaDef{
			Name:             "collect window",
			ShortDescription: "collects items from a stream until released",
			Icon:             "window",
			Tags:             []string{"stream", "window"},
			DocURL:           "https://bitspark.de/slang/docs/operator/window-collect",
		},
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
		DelegateDefs: map[string]*core.DelegateDef{},
		PropertyDefs: core.PropertyMap{
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
