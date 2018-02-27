package builtin

import (
	"slang/core"
)

var aggregateOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		In: core.PortDef{
			Type: "map",
			Map: map[string]*core.PortDef{
				"init": {
					Type:    "generic",
					Generic: "stateType",
				},
				"items": {
					Type: "stream",
					Stream: &core.PortDef{
						Type:    "generic",
						Generic: "itemType",
					},
				},
			},
		},
		Out: core.PortDef{
			Type:    "generic",
			Generic: "stateType",
		},
		Delegates: map[string]*core.DelegateDef{
			"iteration": {
				In: core.PortDef{
					Type: "stream",
					Stream: &core.PortDef{
						Type:    "generic",
						Generic: "stateType",
					},
				},
				Out: core.PortDef{
					Type: "stream",
					Stream: &core.PortDef{
						Type: "map",
						Map: map[string]*core.PortDef{
							"item": {
								Type:    "generic",
								Generic: "itemType",
							},
							"state": {
								Type:    "generic",
								Generic: "stateType",
							},
						},
					},
				},
			},
		},
	},
	oFunc: func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
		iIn := dels["iteration"].In()
		iOut := dels["iteration"].Out()
		for true {
			state := in.Map("init").Pull()

			// Redirect all markers
			if core.IsMarker(state) {
				if !core.IsMarker(in.Map("items").Stream().Pull()) {
					panic("should be marker")
				}
				out.Push(state)
				continue
			}

			in.Map("items").PullBOS()

			iOut.PushBOS()
			iIn.PullBOS()

			for true {
				item := in.Map("items").Stream().Pull()

				if core.IsMarker(item) {
					if in.Map("items").OwnEOS(item) {
						iOut.PushEOS()
						iIn.PullEOS()
						out.Push(state)
						break
					} else {
						panic("unexpected unknown marker")
					}
				}

				iOut.Stream().Map("item").Push(item)
				iOut.Stream().Map("state").Push(state)

				state = iIn.Stream().Pull()
			}
		}
	},
	oConnFunc: func(dest, src *core.Port) error {
		o := dest.Operator()
		if dest == o.In().Map("items") {
			iOut := o.Delegate("iteration").Out()
			iOut.SetStreamSource(src.StreamSource())
		}
		return nil
	},
}
