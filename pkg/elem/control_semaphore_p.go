package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"sync"
)

type semaphoreStore struct {
	semaphore chan bool
}

var semaphoreStores map[string]*semaphoreStore
var semaphoreMutex *sync.Mutex

func getSemaphoreStore(semaphore string) *semaphoreStore {
	semaphoreMutex.Lock()
	semStore, ok := semaphoreStores[semaphore]
	if !ok {
		semStore = &semaphoreStore{
			semaphore: make(chan bool, 1),
		}
		semaphoreStores[semaphore] = semStore
	}
	semaphoreMutex.Unlock()
	return semStore
}

var controlSemaphorePCfg = &builtinConfig{
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

			semStore.semaphore <- true
			out.Push(i)
		}
	},
}
