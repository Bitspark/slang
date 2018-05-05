package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
	"strings"
	"strconv"
)

var serializeOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"el_{indexes}": {
							Type:    "generic",
							Generic: "itemType",
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
		DelegateDefs: map[string]*core.DelegateDef{},
		PropertyDefs: map[string]*core.TypeDef{
			"indexes": {
				Type: "stream",
				Stream: &core.TypeDef{
					Type: "number",
				},
			},
		},
	},
	oFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			im := i.(map[string]interface{})

			maxIndex := -1
			for k := range im {
				if !strings.HasPrefix(k, "el_") {
					panic("malformed entry")
				}
				index, _ := strconv.Atoi(k[3:])
				if index > maxIndex {
					maxIndex = index
				}
			}

			stream := make([]interface{}, maxIndex + 1)
			for k, v := range im {
				index, _ := strconv.Atoi(k[3:])
				stream[index] = v
			}

			out.Push(stream)
		}
	},
}
