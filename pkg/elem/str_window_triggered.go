package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"sync"
)

var streamWindowTriggeredCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"trigger": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type: "trigger",
							},
						},
						"stream": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type:    "generic",
								Generic: "itemType",
							},
						},
					},
				},
				Out: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "stream",
						Stream: &core.TypeDef{
							Type:    "generic",
							Generic: "itemType",
						},
					},
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()

		var mutex = &sync.Mutex{}
		collecting := false
		var window []interface{}

		go func() {
			for !op.CheckStop() {
				i := in.Map("trigger").Stream().Pull()
				if !in.Map("trigger").OwnBOS(i) {
					out.Push(i)
					continue
				}
				mutex.Lock()
				collecting = true
				mutex.Unlock()

				out.PushBOS()
				for {
					i := in.Map("trigger").Stream().Pull()
					if in.Map("trigger").OwnEOS(i) {
						mutex.Lock()
						window = nil
						collecting = false
						mutex.Unlock()
						break
					}

					mutex.Lock()
					out.Stream().Push(window)
					window = nil
					mutex.Unlock()
				}
				out.PushEOS()
			}
		}()

		for !op.CheckStop() {
			i := in.Map("stream").Stream().Pull()
			if core.IsMarker(i) {
				continue
			}

			mutex.Lock()
			if collecting {
				window = append(window, i)
			}
			mutex.Unlock()
		}
	},
	opConnFunc: func(op *core.Operator, dst, src *core.Port) error {
		if dst == op.Main().In().Map("trigger") {
			op.Main().Out().SetStreamSource(src.StreamSource())
		}
		return nil
	},
}
