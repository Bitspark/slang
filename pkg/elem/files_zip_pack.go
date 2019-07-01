package elem

import (
	"archive/zip"
	"bytes"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var filesZIPPackCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: uuid.MustParse("dc5325bc-a816-47c8-8a8a-f741497459f7"),
		Meta: core.OperatorMetaDef{
			Name:             "pack ZIP",
			ShortDescription: "packs a stream of binary content into a zip archive",
			Icon:             "file-archive",
			Tags:             []string{"zip", "compression"},
			DocURL:           "https://bitspark.de/slang/docs/operator/pack-zip",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "map",
						Map: map[string]*core.TypeDef{
							"path": {
								Type: "string",
							},
							"file": {
								Type: "binary",
							},
						},
					},
				},
				Out: core.TypeDef{
					Type: "binary",
				},
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			i := in.Stream().Pull()
			if !in.OwnBOS(i) {
				out.Push(i)
				continue
			}

			buf := new(bytes.Buffer)
			zipWriter := zip.NewWriter(buf)

			for {
				i = in.Pull()
				if in.OwnEOS(i) {
					break
				}

				im := i.(map[string]interface{})

				path := im["path"].(string)
				file := im["file"].(core.Binary)

				fileWriter, _ := zipWriter.Create(path)
				fileWriter.Write(file)
			}

			zipWriter.Close()

			out.Push(core.Binary(buf.Bytes()))
		}
	},
}
