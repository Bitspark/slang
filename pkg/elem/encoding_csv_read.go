package elem

import (
	"encoding/csv"
	"io"
	"strings"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
	"github.com/thoas/go-funk"
)

var encodingCSVReadId = uuid.MustParse("77d60459-f8b5-4f4b-b293-740164c49a82")
var encodingCSVReadCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: encodingCSVReadId,
		Meta: core.BlueprintMetaDef{
			Name:             "read CSV",
			ShortDescription: "reads a CSV file and emits a stream of lines, separated into columns",
			Icon:             "file-csv",
			Tags:             []string{"file"},
			DocURL:           "https://bitspark.de/slang/docs/operator/read-csv",
		},
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
		DelegateDefs: map[string]*core.DelegateDef{},
		PropertyDefs: core.PropertyMap{
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
	opFunc: func(op *core.Operator) {
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
			colNames := op.Property("columns").([]interface{})

			mapSize := outStream.MapLength()

			r := csv.NewReader(strings.NewReader(csvText))
			r.Comma = rune(op.Property("delimiter").(string)[0])

			for {
				rec, err := r.Read()
				if err == io.EOF {
					break
				}
				if len(rec) < mapSize {
					break
				}

				if mapping == nil {
					mapping = rec
				} else {
					for i, col := range mapping {
						if funk.Contains(colNames, col) {
							outStream.Map("col_" + col).Push(rec[i])
						}
					}
				}
			}
			out.PushEOS()
		}
	},
	opConnFunc: func(op *core.Operator, dst, src *core.Port) error {
		return nil
	},
}
