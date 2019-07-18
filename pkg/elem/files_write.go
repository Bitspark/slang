package elem

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var filesWriteCfg = &builtinConfig{
	blueprint: core.Blueprint{
		Id: uuid.MustParse("9b61597d-cfbc-42d1-9620-210081244ba1"),
		Meta: core.BlueprintMetaDef{
			Name:             "write file",
			ShortDescription: "creates or replaces a file and writes binary data to it",
			Icon:             "file-signature",
			Tags:             []string{"file"},
			DocURL:           "https://bitspark.de/slang/docs/operator/write-file",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"content": {
							Type: "binary",
						},
						"filename": {
							Type: "string",
						},
					},
				},
				Out: core.TypeDef{
					Type: "string",
				},
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			data := i.(map[string]interface{})
			var content []byte
			if b, ok := data["content"].(core.Binary); ok {
				content = b
			}
			if s, ok := data["content"].(string); ok {
				content = []byte(s)
			}
			filename := data["filename"].(string)

			err := ioutil.WriteFile(filepath.Clean(filename), content, os.ModePerm)

			if err == nil {
				out.Push(nil)
			} else {
				out.Push(err.Error())
			}
		}
	},
}
