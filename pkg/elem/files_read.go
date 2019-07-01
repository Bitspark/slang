package elem

import (
	"io/ioutil"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var filesReadId = uuid.MustParse("f7eecf2c-6504-478f-b2fa-809bec71463c")
var filesReadCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: filesReadId,
		Meta: core.OperatorMetaDef{
			Name:             "read file",
			ShortDescription: "reads the contents of a file and emits them",
			Icon:             "file",
			Tags:             []string{"file"},
			DocURL:           "https://bitspark.de/slang/docs/operator/read-file",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "string",
				},
				Out: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"content": {
							Type: "binary",
						},
						"error": {
							Type: "string",
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
			file, marker := in.PullString()
			if marker != nil {
				out.Push(marker)
				continue
			}

			path := filepath.Clean(file)
			if strings.HasPrefix(path, "~") {
				usr, _ := user.Current()
				dir := usr.HomeDir
				path = filepath.Join(dir, path[1:])
			}
			content, err := ioutil.ReadFile(path)
			if err != nil {
				out.Map("content").Push(nil)
				out.Map("error").Push(err.Error())
				continue
			}

			out.Map("content").Push(core.Binary(content))
			out.Map("error").Push(nil)
		}
	},
}
