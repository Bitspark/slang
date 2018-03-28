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
						Type:    "generic",
						Generic: "colMap",
					},
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{
		},
	},
	oFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for {
			csvText, marker := in.PullString()
			if marker != nil {
				out.Push(marker)
				continue
			}

			outStream := out.Stream()

			out.PushBOS()

			mapping, _ := op.Property("colMapping").([]string)
			mapSize := outStream.MapSize()

			r := csv.NewReader(strings.NewReader(csvText))
			r.Comma = op.Property("delimiter").(rune)
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
						outStream.Map(col).Push(rec[i])
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
