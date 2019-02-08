package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/tealeg/xlsx"
)

var encodingXLSXReadCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Meta: core.OperatorMetaDef{
			Name: "read Excel",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "binary",
				},
				Out: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "map",
						Map: map[string]*core.TypeDef{
							"name": {
								Type: "string",
							},
							"table": {
								Type: "stream",
								Stream: &core.TypeDef{
									Type: "stream",
									Stream: &core.TypeDef{
										Type: "primitive",
									},
								},
							},
						},
					},
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			b, i := in.PullBinary()
			if i != nil {
				out.Push(i)
				continue
			}
			xlsxFile, err := xlsx.OpenBinary(b)
			if err != nil {
				panic(err)
			}
			out.PushBOS()
			outTable := out.Stream().Map("table")
			for _, sheet := range xlsxFile.Sheets {
				out.Stream().Map("name").Push(sheet.Name)
				outTable.PushBOS()
				for _, row := range sheet.Rows {
					outTable.Stream().PushBOS()
					for _, col := range row.Cells {
						outTable.Stream().Stream().Push(col.Value)
					}
					outTable.Stream().PushEOS()
				}
				outTable.PushEOS()
			}
			out.PushEOS()
		}
	},
}
