package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
	"sync"
)

var windowTriggeredOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		Services: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.PortDef{
					Type: "map",
					Map: map[string]*core.PortDef{
						"trigger": {
							Type: "stream",
							Stream: &core.PortDef{
								Type: "trigger",
							},
						},
						"stream": {
							Type: "stream",
							Stream: &core.PortDef{
								Type:    "generic",
								Generic: "itemType",
							},
						},
					},
				},
				Out: core.PortDef{
					Type: "stream",
					Stream: &core.PortDef{
						Type: "stream",
						Stream: &core.PortDef{
							Type:    "generic",
							Generic: "itemType",
						},
					},
				},
			},
		},
		Delegates: map[string]*core.DelegateDef{
		},
	},
	oFunc: func(srvs map[string]*core.Service, dels map[string]*core.Delegate, store interface{}) {
		in := srvs[core.MAIN_SERVICE].In()
		out := srvs[core.MAIN_SERVICE].Out()

		var mutex = &sync.Mutex{}
		collecting := false
		var window []interface{}

		go func() {
			for {
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

		for {
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
	oPropFunc: func(op *core.Operator, i map[string]interface{}) error {
		return nil
	},
	oConnFunc: func(dest, src *core.Port) error {
		o := dest.Operator()
		if dest == o.Main().In().Map("trigger") {
			o.Main().Out().SetStreamSource(src.StreamSource())
		}
		return nil
	},
}
