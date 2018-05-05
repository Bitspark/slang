package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
)

var concatOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"stream_{streams}": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type: "generic",
								Generic: "itemType",
							},
						},
					},
				},
				Out: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "generic",
						Generic: "itemType",
					},
				},
			},
		},
		PropertyDefs: map[string]*core.TypeDef{
			"streams": {
				Type: "stream",
				Stream: &core.TypeDef{
					Type: "string",
				},
			},
		},
	},
	oFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		indexesProp := op.Property("streams").([]interface{})
		streams := make([]*core.Port, len(indexesProp))
		for i, idxProp := range indexesProp {
			streams[i] = in.Map("stream_"+idxProp.(string))
		}
		for {
			item := streams[0].Stream().Pull()
			if !streams[0].OwnBOS(item) {
				for i := 1; i < len(streams); i++ {
					streams[i].Stream().Pull()
				}
				out.Push(item)
				continue
			}

			out.PushBOS()
			for i := 0; i < len(streams); i++ {
				for {
					item = streams[i].Stream().Pull()
					if streams[i].OwnEOS(item) {
						if i+1 < len(streams) {
							streams[i+1].PullBOS()
						}
						break
					}
					out.Stream().Push(item)
				}
			}
			out.PushEOS()
		}
	},
}
