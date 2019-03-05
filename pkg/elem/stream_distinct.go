package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"strconv"
)

var streamDistinctCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: "d8914bae-2878-46f3-b468-9e7faea7a463",
		Meta: core.OperatorMetaDef{
			Name: "distinct values",
			ShortDescription: "takes a streams and emits distinct items",
			Icon: "shapes",
			Tags: []string{"stream"},
			DocURL: "https://bitspark.de/slang/docs/operator/distinct",
		},
		ServiceDefs: map[string]*core.ServiceDef{
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
						Type:    "generic",
						Generic: "itemType",
					},
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{
			"checker": {
				Out: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"a": {
							Type:    "generic",
							Generic: "itemType",
						},
						"b": {
							Type:    "generic",
							Generic: "itemType",
						},
					},
				},
				In: core.TypeDef{
					Type: "boolean",
				},
			},
			"hasher": {
				Out: core.TypeDef{
					Type:    "generic",
					Generic: "itemType",
				},
				In: core.TypeDef{
					Type: "string",
				},
			},
		},
		PropertyDefs: core.TypeDefMap{},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		inStream := in.Stream()
		out := op.Main().Out()
		outStream := out.Stream()

		hasher := op.Delegate("hasher")
		checker := op.Delegate("checker")

		for !op.CheckStop() {
			i := inStream.Pull()
			if !in.OwnBOS(i) {
				out.Push(i)
				continue
			}

			m := make(map[string]interface{})

			for {
				i = inStream.Pull()
				if in.OwnEOS(i) {
					break
				}

				hasher.Out().Push(i)
				h := hasher.In().Pull().(string)
				num := 0

				for {
					hn := h + strconv.Itoa(num)
					if mi, ok := m[hn]; ok {
						checker.Out().Map("a").Push(i)
						checker.Out().Map("b").Push(mi)

						if checker.In().Pull().(bool) {
							m[hn] = i
							break
						} else {
							num++
						}
					} else {
						m[hn] = i
						break
					}
				}
			}

			out.PushBOS()
			for _, mi := range m {
				outStream.Push(mi)
			}
			out.PushEOS()
		}
	},
}
