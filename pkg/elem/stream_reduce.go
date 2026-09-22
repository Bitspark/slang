package elem

import (
	"sync"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var streamReduceCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: uuid.MustParse("b95e6da8-9770-4a04-a73d-cdfe2081870f"),
		Meta: core.BlueprintMetaDef{
			Name:             "reduce",
			ShortDescription: "reduces items of a stream pairwise using a reducer delegate",
			Icon:             "compress-alt",
			Tags:             []string{"stream"},
			DocURL:           "https://bitspark.de/slang/docs/operator/reduce",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type:    "generic",
						Generic: "itemType",
					},
				},
				Out: core.TypeDef{
					Type:    "generic",
					Generic: "itemType",
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{
			"reducer": {
				In: core.TypeDef{
					Type:    "generic",
					Generic: "itemType",
				},
				Out: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"a": {
							Type:    "generic",
							Generic: "itemType",
						},
						"b": {
							Type:    "generic",
							Generic: "itemType",
						},
					},
				},
			},
		},
		PropertyDefs: core.PropertyMap{
			"emptyValue": {
				Type:    "generic",
				Generic: "itemType",
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		sIn := op.Delegate("reducer").In()
		sOut := op.Delegate("reducer").Out()
		nullValue := op.Property("emptyValue")
		for !op.CheckStop() {
			i := in.Stream().Pull()

			if !in.OwnBOS(i) {
				out.Push(i)
				continue
			}

			result, ok := reduceStream(op, in, sIn, sOut, nullValue)
			if !ok {
				return
			}
			out.Push(result)
		}
	},
}

// Read the source concurrently with the delegate, but never hold the pool lock
// during port I/O. Both workers must be able to finish if either input closes.
func reduceStream(op *core.Operator, in, sIn, sOut *core.Port, empty interface{}) (interface{}, bool) {
	var mutex sync.Mutex
	ready := sync.NewCond(&mutex)
	pool := []interface{}{}
	done := false
	completed := false
	finished := make(chan struct{})
	finishInput := func() {
		mutex.Lock()
		done = true
		ready.Broadcast()
		mutex.Unlock()
	}
	defer finishInput()
	op.Go(func() {
		defer close(finished)
		for {
			mutex.Lock()
			for len(pool) < 2 && !done {
				ready.Wait()
			}
			if len(pool) < 2 {
				completed = true
				mutex.Unlock()
				return
			}
			a, b := pool[0], pool[1]
			pool = pool[2:]
			mutex.Unlock()
			sOut.Push(map[string]interface{}{"a": a, "b": b})
			item := sIn.Pull()
			mutex.Lock()
			// Prepend the reduction to preserve source order.
			pool = append([]interface{}{item}, pool...)
			mutex.Unlock()
		}
	})
	for {
		item := in.Stream().Pull()
		if in.OwnEOS(item) {
			break
		}
		mutex.Lock()
		pool = append(pool, item)
		ready.Signal()
		mutex.Unlock()
	}
	finishInput()
	select {
	case <-finished:
	case <-op.Done():
		return nil, false
	}
	mutex.Lock()
	defer mutex.Unlock()
	if !completed {
		return nil, false
	}
	if len(pool) == 1 {
		return pool[0], true
	}
	return empty, true
}
