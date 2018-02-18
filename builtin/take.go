package builtin

import (
	"slang/core"
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
				"select": {
					Type: "stream",
					Stream: &core.PortDef{
						Type: "boolean",
					},
				},
			},
		},
		Out: core.PortDef{
			Type: "map",
			Map: map[string]*core.PortDef{
				"result": {
					Type: "stream",
					Stream: &core.PortDef{
						Type:    "generic",
						Generic: "itemType",
					},
				},
				"compare": {
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
		for true {
			t := in.Map("true").Stream().Pull()
			f := in.Map("false").Stream().Pull()

			if !in.Map("true").OwnBOS(t) {
				if core.IsMarker(t) {
					if t == f {
						out.Map("compare").Stream().Push(t)
						sel := in.Map("select").Stream().Pull()

						if sel != t {
							panic("expected marker")
						}

						out.Map("result").Stream().Push(t)

						continue
					}
					panic("expected equal marker")
				}
				panic("expected marker")
			}

			if !in.Map("false").OwnBOS(f) {
				panic("expected BOS")
			}

			out.Map("result").PushBOS()

			out.Map("compare").PushBOS()
			sel := in.Map("select").Stream().Pull()
			if !in.Map("select").OwnBOS(sel) {
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
						out.Map("result").Stream().Push(f)
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
						out.Map("result").Stream().Push(t)
						t = in.Map("true").Stream().Pull()
					}
				} else {
					// Send to comparator
					out.Map("compare").Stream().Map("true").Push(t)
					out.Map("compare").Stream().Map("false").Push(f)

					// Get result
					s, ok := in.Map("select").Stream().Pull().(bool)

					if !ok {
						panic("expected boolean")
					}

					if s {
						out.Map("result").Stream().Push(t)
						t = nil
					} else {
						out.Map("result").Stream().Push(f)
						f = nil
					}
				}
			}

		end:
			out.Map("compare").PushEOS()
			sel = in.Map("select").Stream().Pull()
			if !in.Map("select").OwnEOS(sel) {
				panic("expected EOS")
			}

			out.Map("result").PushEOS()
		}
	},
}
