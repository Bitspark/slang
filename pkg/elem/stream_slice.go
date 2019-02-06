package elem

import (
	"github.com/Bitspark/slang/pkg/core"
)

var streamSliceCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"offset": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type:    "number",
							},
						},
						"length": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type:    "number",
							},
						},
						"step": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type:    "number",
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
					Type: "map",
					Map: map[string]*core.TypeDef{
						"stream": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type:    "generic",
								Generic: "itemType",
							},
						},
					},
				},
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			im := i.(map[string]interface{})

			stream := im["stream"].([]interface{})
			offset := int(im["offset"].(float64))
			length := int(im["length"].(float64))
			step := int(im["step"].(float64))

			until := len(stream)
			if until > offset + length {
				until = offset + length
			}

			out.Map("stream").PushBOS()
			outStream := out.Map("stream").Stream()
			for i := offset; i < until; i += step {
				outStream.Push(stream[i])
			}
			out.Map("stream").PushEOS()
		}
	},
}
