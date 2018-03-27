package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
)

var windowCountOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		Services: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type:    "generic",
						Generic: "itemType",
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
		Delegates: map[string]*core.DelegateDef{
		},
	},
	oFunc: func(op *core.Operator) {
		s := struct {
			size  int
			slide int
			start int
			end   int
		}{
			op.Property("size").(int),
			op.Property("slide").(int),
			op.Property("start").(int),
			op.Property("end").(int),
		}
		in := op.Main().In()
		out := op.Main().Out()
		for {
			i := in.Stream().Pull()
			if core.IsMarker(i) && !in.OwnBOS(i) {
				out.Push(i)
				continue
			}

			out.PushBOS()
			var window []interface{}
			rest := s.start
			for {
				i := in.Stream().Pull()
				if core.IsMarker(i) && in.OwnEOS(i) {
					if len(window) >= s.end {
						out.Stream().Push(window)
					}
					break
				}

				window = append(window, i)
				rest--

				if len(window) > s.size {
					window = window[1:]
				}
				if rest == 0 {
					out.Stream().Push(window)
					if s.slide >= s.size {
						window = window[:0]
					} else {
						window = window[s.slide-(s.size-len(window)):]
					}
					rest = s.slide
				}
			}
			out.PushEOS()
		}
	},
	oPropFunc: func(props core.Properties) error {
		props["size"] = int(props["size"].(float64))
		props["slide"] = int(props["slide"].(float64))
		props["start"] = int(props["start"].(float64))
		props["end"] = int(props["end"].(float64))

		return nil
	},
	oConnFunc: func(op *core.Operator, dst, src *core.Port) error {
		return nil
	},
}
