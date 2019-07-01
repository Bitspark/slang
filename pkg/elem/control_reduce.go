package elem

import (
	"sync"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var controlReduceId = uuid.MustParse("b95e6da8-9770-4a04-a73d-cdfe2081870f")
var controlReduceCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: controlReduceId,
		Meta: core.OperatorMetaDef{
			Name:             "reduce",
			ShortDescription: "reduces the items of a stream pairwise using a reducer delegate",
			Icon:             "compress-alt",
			Tags:             []string{"data", "stream"},
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
		PropertyDefs: map[string]*core.TypeDef{
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

			mutex := &sync.Mutex{}
			pool := []interface{}{}
			done := false
			doneChan := make(chan bool)

			// Reducer
			go func() {
				for {
					mutex.Lock()
					if done && len(pool) < 2 {
						doneChan <- true
						break
					}
					mutex.Unlock()

					mutex.Lock()
					if len(pool) > 1 {
						sOut.Push(map[string]interface{}{"a": pool[0], "b": pool[1]})
						pool = pool[2:]
					} else {
						mutex.Unlock()
						continue
					}
					mutex.Unlock()

					i := sIn.Pull()

					mutex.Lock()
					pool = append(pool, i)
					mutex.Unlock()
				}
			}()

			for {
				// Stream items

				i = in.Stream().Pull()
				if in.OwnEOS(i) {
					done = true
					break
				}

				mutex.Lock()
				pool = append(pool, i)
				mutex.Unlock()
			}

			<-doneChan

			if len(pool) == 1 {
				out.Push(pool[0])
			} else {
				out.Push(nullValue)
			}
		}
	},
}
