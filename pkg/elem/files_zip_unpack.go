package elem

import (
	"archive/zip"
	"bytes"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var filesZIPUnpackCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: uuid.MustParse("04714d4a-1d5d-4b68-b614-524dd4662ef4"),
		Meta: core.BlueprintMetaDef{
			Name:             "unpack ZIP",
			ShortDescription: "unpacks a zip archive and emits a stream of files and their binary content",
			Icon:             "file-archive",
			Tags:             []string{"zip", "compression"},
			DocURL:           "https://bitspark.de/slang/docs/operator/unpack-zip",
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
							"path": {
								Type: "string",
							},
							"file": {
								Type: "binary",
							},
						},
					},
				},
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) || i == nil {
				out.Push(i)
				continue
			}

			b := i.(core.Binary)
			reader := bytes.NewReader(b)
			zipReader, err := zip.NewReader(reader, reader.Size())
			if err != nil {
				out.Push(nil)
				continue
			}

			out.PushBOS()
			for _, file := range zipReader.File {
				out.Stream().Map("path").Push(file.Name)
				fileReader, _ := file.Open()
				buf := new(bytes.Buffer)
				buf.ReadFrom(fileReader)
				out.Stream().Map("file").Push(core.Binary(buf.Bytes()))
			}
			out.PushEOS()
		}
	},
}
