package elem

import (
	"sync"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
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

var controlSemaphorePId = uuid.MustParse("199f14c3-3e25-4813-aaba-7ec7fa3d94e2")
var controlSemaphorePCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: controlSemaphorePId,
		Meta: core.OperatorMetaDef{
			Name:             "semaphore P",
			ShortDescription: "tries to acquire semaphore token",
			Icon:             "traffic-light-stop",
			Tags:             []string{"control", "sync"},
			DocURL:           "https://bitspark.de/slang/docs/operator/semaphore-p",
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
