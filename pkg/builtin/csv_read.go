package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
	"strings"
	"encoding/csv"
	"io"
)

var csvReadOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "string",
				},
				Out: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "map",
						Map: map[string]*core.TypeDef{
							"col_{columns}": {
								Type: "string",
							},
						},
					},
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{
		},
		PropertyDefs: map[string]*core.TypeDef{
			"delimiter": {
				Type: "string",
			},
			"columns": {
				Type: "stream",
				Stream: &core.TypeDef{
					Type: "string",
				},
			},
		},
	},
	oFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			csvText, marker := in.PullString()
			if marker != nil {
				out.Push(marker)
				continue
			}

			outStream := out.Stream()

			out.PushBOS()

			var mapping []string
			for _, col := range op.Property("columns").([]interface{}) {
				mapping = append(mapping, col.(string))
			}
			mapSize := outStream.MapSize()

			r := csv.NewReader(strings.NewReader(csvText))
			r.Comma = rune(op.Property("delimiter").(string)[0])
			for {
				rec, err := r.Read()
				if err == io.EOF {
					break
				}
				if len(rec) != mapSize {
					break
				}

				if mapping == nil {
					mapping = rec
				} else {
					for i, col := range mapping {
						outStream.Map("col_" + col).Push(rec[i])
					}
				}
			}
			out.PushEOS()
		}
	},
	oConnFunc: func(op *core.Operator, dst, src *core.Port) error {
		return nil
	},
}
