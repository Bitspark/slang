package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"io/ioutil"
	"os"
	"path/filepath"
)

var filesWriteCfg = &builtinConfig{
	opDef: core.OperatorDef{
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
			if b, ok := data["content"].([]byte); ok {
				content = b
			}
			if s, ok := data["content"].(string); ok {
				content = []byte(s)
			}
			filename := data["filename"].(string)

			err := ioutil.WriteFile(filepath.Join(core.WORKING_DIR, filename), content, os.ModePerm)

			if err == nil {
				out.Push(nil)
			} else {
				out.Push(err.Error())
			}
		}
	},
}
