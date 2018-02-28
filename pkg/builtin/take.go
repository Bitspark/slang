package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
)

var takeOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		In: core.PortDef{
			Type: "map",
			Map: map[string]*core.PortDef{
				"true": {
					Type: "stream",
					Stream: &core.PortDef{
						Type:    "generic",
						Generic: "itemType",
					},
				},
				"false": {
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
				Type:    "generic",
				Generic: "itemType",
			},
		},
		Delegates: map[string]*core.DelegateDef{
			"compare": {
				In: core.PortDef{
					Type: "stream",
					Stream: &core.PortDef{
						Type: "boolean",
					},
				},
				Out: core.PortDef{
					Type: "stream",
					Stream: &core.PortDef{
						Type: "map",
						Map: map[string]*core.PortDef{
							"true": {
								Type:    "generic",
								Generic: "itemType",
							},
							"false": {
								Type:    "generic",
								Generic: "itemType",
							},
						},
					},
				},
			},
		},
	},
	oFunc: func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
		cIn := dels["compare"].In()
		cOut := dels["compare"].Out()
		for true {
			t := in.Map("true").Stream().Pull()
			f := in.Map("false").Stream().Pull()

			if !in.Map("true").OwnBOS(t) {
				if core.IsMarker(t) {
					if t == f {
						cOut.Stream().Push(t)
						sel := cIn.Stream().Pull()

						if sel != t {
							panic("expected marker")
						}

						out.Stream().Push(t)

						continue
					}
					panic("expected equal marker")
				}
				panic("expected marker")
			}

			if !in.Map("false").OwnBOS(f) {
				panic("expected BOS")
			}

			out.PushBOS()

			cOut.PushBOS()
			sel := cIn.Stream().Pull()
			if !cIn.OwnBOS(sel) {
				panic("expected BOS")
			}

			t = nil
			f = nil
			for true {
				if t == nil {
					t = in.Map("true").Stream().Pull()
				}
				if f == nil {
					f = in.Map("false").Stream().Pull()
				}

				if core.IsMarker(t) {
					if !in.Map("true").OwnEOS(t) {
						panic("expected EOS")
					}
					for true {
						if core.IsMarker(f) {
							if !in.Map("false").OwnEOS(f) {
								panic("expected EOS")
							}
							goto end
						}
						out.Stream().Push(f)
						f = in.Map("false").Stream().Pull()
					}
				} else if core.IsMarker(f) {
					if !in.Map("false").OwnEOS(f) {
						panic("expected EOS")
					}
					for true {
						if core.IsMarker(t) {
							if !in.Map("true").OwnEOS(t) {
								panic("expected EOS")
							}
							goto end
						}
						out.Stream().Push(t)
						t = in.Map("true").Stream().Pull()
					}
				} else {
					// Send to comparator
					cOut.Stream().Map("true").Push(t)
					cOut.Stream().Map("false").Push(f)

					// Get result
					s, ok := cIn.Stream().Pull().(bool)

					if !ok {
						panic("expected boolean")
					}

					if s {
						out.Stream().Push(t)
						t = nil
					} else {
						out.Stream().Push(f)
						f = nil
					}
				}
			}

		end:
			cOut.PushEOS()
			sel = cIn.Stream().Pull()
			if !cIn.OwnEOS(sel) {
				panic("expected EOS")
			}

			out.PushEOS()
		}
	},
}
