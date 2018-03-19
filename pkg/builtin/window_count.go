package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
)

type windowCountStore struct {
	size  int
	slide int
	start int
	end   int
}

var windowCountOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		In: core.PortDef{
			Type:    "generic",
			Generic: "inStreams",
		},
		Out: core.PortDef{
			Type: "stream",
			Stream: &core.PortDef{
				Type: "stream",
				Stream: &core.PortDef{
					Type:    "generic",
					Generic: "outStreams",
				},
			},
		},
		Delegates: map[string]*core.DelegateDef{
		},
	},
	oFunc: func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
		s := store.(windowCountStore)
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
	oPropFunc: func(op *core.Operator, i map[string]interface{}) error {
		store := windowCountStore{}

		store.size = int(i["size"].(float64))
		store.slide = int(i["slide"].(float64))
		store.start = int(i["start"].(float64))
		store.end = int(i["end"].(float64))

		op.SetStore(store)
		return nil
	},
	oConnFunc: func(dest, src *core.Port) error {
		return nil
	},
}
