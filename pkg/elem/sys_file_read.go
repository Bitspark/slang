package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"io/ioutil"
	"path/filepath"
	"github.com/Bitspark/slang/pkg/utils"
)

var fileReadOpCfg = &builtinConfig{
	opDef: core.OperatorDef{
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

			content, err := ioutil.ReadFile(filepath.Join(core.WORKING_DIR, file))
			if err != nil {
				out.Map("content").Push(nil)
				out.Map("error").Push(err.Error())
				continue
			}

			out.Map("content").Push(utils.Binary(content))
			out.Map("error").Push(nil)
		}
	},
}
