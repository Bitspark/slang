package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
	"sync"
)

var windowTriggeredOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
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
		Delegates: map[string]*core.DelegateDef{
		},
	},
	oFunc: func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
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
		if dest == o.In().Map("trigger") {
			o.Out().SetStreamSource(src.StreamSource())
		}
		return nil
	},
}
