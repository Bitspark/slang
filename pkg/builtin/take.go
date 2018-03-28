package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
)

var takeOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"true": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type:    "generic",
								Generic: "itemType",
							},
						},
						"false": {
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
						Type:    "generic",
						Generic: "itemType",
					},
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{
			"compare": {
				In: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "boolean",
					},
				},
				Out: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "map",
						Map: map[string]*core.TypeDef{
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
	oFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		cIn := op.Delegate("compare").In()
		cOut := op.Delegate("compare").Out()
		for {
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
			cIn.PullBOS()

			t = nil
			f = nil
			for {
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
					for {
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
					for {
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
			cIn.PullEOS()

			out.PushEOS()
		}
	},
}
