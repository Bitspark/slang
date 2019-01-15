package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/utils"
	"os"
	"path/filepath"
)

var filesAppendCfg = &builtinConfig{
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
		PropertyDefs: map[string]*core.TypeDef{
			"newLine": {
				Type: "boolean",
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		newLine := op.Property("newLine").(bool)
		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			data := i.(map[string]interface{})
			var content []byte
			if b, ok := data["content"].(utils.Binary); ok {
				content = b
			}
			if s, ok := data["content"].(string); ok {
				content = []byte(s)
			}
			filename := data["filename"].(string)

			f, err := os.OpenFile(filepath.Clean(filename), os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				out.Push(err.Error())
				continue
			}

			_, err = f.Write(content)
			if err != nil {
				out.Push(err.Error())
				continue
			}

			if newLine {
				_, err = f.Write([]byte("\n"))
				if err != nil {
					out.Push(err.Error())
					continue
				}
			}

			out.Push(nil)
		}
	},
}
