package elem

import (
	"bytes"
	"fmt"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/tealeg/xlsx"
	"strconv"
)

var encodingXLSXWriteCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: "9accc6f6-9e3b-45ba-9128-6040e1279655",
		Meta: core.OperatorMetaDef{
			Name: "write Excel",
			ShortDescription: "encodes Excel data from a stream of sheets, each being a 2d-stream of cells",
			Icon: "file-excel",
			Tags: []string{"excel", "encoding"},
			DocURL: "https://bitspark.de/slang/docs/operator/write-excel",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
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
				Out: core.TypeDef{
					Type: "binary",
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			data := in.Pull()
			if core.IsMarker(data) {
				out.Push(data)
				continue
			}

			xlsxFile := xlsx.NewFile()

			sheetsData := data.([]interface{})

			for _, sheetData := range sheetsData {
				sheetMap := sheetData.(map[string]interface{})
				sheet := &xlsx.Sheet{
					Name: sheetMap["name"].(string),
				}
				tableData := sheetMap["table"].([]interface{})
				for _, tableRowData := range tableData {
					row := sheet.AddRow()
					tableRow := tableRowData.([]interface{})
					for _, tableCellData := range tableRow {
						cell := row.AddCell()

						if tableCellStr, ok := tableCellData.(string); ok {
							cell.Value = tableCellStr
						} else if tableCellNum, ok := tableCellData.(float64); ok {
							cell.Value = strconv.FormatFloat(tableCellNum, 'f', -1, 64)
						} else {
							cell.Value = fmt.Sprint("%v", tableCellData)
						}
					}
				}
				xlsxFile.Sheets = append(xlsxFile.Sheets, sheet)
			}

			bts := make([]byte, 0)
			b := bytes.NewBuffer(bts)
			xlsxFile.Write(b)

			out.Push(core.Binary(b.Bytes()))
		}
	},
}
