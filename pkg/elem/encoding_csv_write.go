package elem

import (
	"bytes"
	"encoding/csv"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var encodingCSVWriteId = uuid.MustParse("fdd1e8e5-6959-4511-bf44-54c1bcbebc12")
var encodingCSVWriteCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: encodingCSVWriteId,
		Meta: core.BlueprintMetaDef{
			Name:             "write CSV",
			ShortDescription: "encodes streams into a single string",
			Icon:             "file-csv",
			Tags:             []string{"file"},
			DocURL:           "https://bitspark.de/slang/docs/operator/write-csv",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				Out: core.TypeDef{
					Type: "string",
				},
				In: core.TypeDef{
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
		DelegateDefs: map[string]*core.DelegateDef{},
		PropertyDefs: core.PropertyMap{
			"includeHeader": {
				TypeDef: core.TypeDef{
					Type: "boolean",
				},
			},
			"delimiter": {
				TypeDef: core.TypeDef{
					Type: "string",
				},
			},
			"columns": {
				TypeDef: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "string",
					},
				},
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			rows := in.Pull()
			if core.IsMarker(rows) {
				out.Push(rows)
				continue
			}

			colNames := op.Property("columns").([]interface{})
			includeHeader := op.Property("includeHeader").(bool)
			var buf bytes.Buffer
			writer := csv.NewWriter(&buf)
			writer.Comma = rune(op.Property("delimiter").(string)[0])

			if includeHeader {
				header := []string{}
				for _, c := range colNames {
					header = append(header, c.(string))
				}
				writer.Write(header)
			}

			for _, r := range rows.([]interface{}) {
				cells := r.(map[string]interface{})
				record := []string{}
				for _, c := range colNames {
					record = append(record, cells["col_"+c.(string)].(string))
				}
				writer.Write(record)
			}
			writer.Flush()
			out.Push(buf.String())
		}
	},
	opConnFunc: func(op *core.Operator, dst, src *core.Port) error {
		return nil
	},
}
