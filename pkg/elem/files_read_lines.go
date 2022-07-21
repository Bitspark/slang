package elem

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var filesReadLinesCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: uuid.MustParse("6124cd6b-5c23-4e17-a714-458d0f8ac1a7"),
		Meta: core.BlueprintMetaDef{
			Name:             "lines from file",
			ShortDescription: "reads the contents of a file line by line and emits them as stream",
			Icon:             "file",
			Tags:             []string{"file"},
			DocURL:           "https://bitspark.de/slang/docs/operator/lines-from-file",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"filename": {
							Type: "string",
						},
					},
				},
				Out: core.TypeDef{
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
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			data := i.(map[string]interface{})
			filename := data["filename"].(string)

			f, err := os.Open(filepath.Clean(filename))
			if err != nil {
				f.Close()
				out.Push(err.Error())
				continue
			}

			buf := bufio.NewReader(f)

			out.PushBOS()
			for line, _, err := buf.ReadLine(); err == nil; line, _, err = buf.ReadLine() {
				out.Stream().Push(string(line))
			}
			out.PushEOS()

			f.Close()
		}
	},
}
