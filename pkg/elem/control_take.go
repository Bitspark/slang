package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var controlTakeId = uuid.MustParse("9bebc4bf-d512-4944-bcb1-5b2c3d5b5471")
var controlTakeCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: controlTakeId,
		Meta: core.OperatorMetaDef{
			Name:             "take",
			ShortDescription: "merges two streams using a compare delegate deciding which item has precedence",
			Icon:             "hand-point-up",
			Tags:             []string{"control"},
			DocURL:           "https://bitspark.de/slang/docs/operator/take",
		},
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
					Type: "boolean",
				},
				Out: core.TypeDef{
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
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		cIn := op.Delegate("compare").In()
		cOut := op.Delegate("compare").Out()
		for !op.CheckStop() {
			t := in.Map("true").Stream().Pull()
			f := in.Map("false").Stream().Pull()

			if !in.Map("true").OwnBOS(t) {
				if core.IsMarker(t) {
					if t == f {
						cOut.Push(t)
						sel := cIn.Pull()

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
					cOut.Map("true").Push(t)
					cOut.Map("false").Push(f)

					// Get result
					s, ok := cIn.Pull().(bool)

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
			out.PushEOS()
		}
	},
}
