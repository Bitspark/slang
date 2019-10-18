package elem

import (
	"github.com/Bitspark/slang/pkg/log"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var systemLogId = uuid.MustParse("8f9c02df-da41-4266-b486-0c22173a6383")
var systemLogCfg = &builtinConfig{
	blueprint: core.Blueprint{
		Id: systemLogId,
		Meta: core.BlueprintMetaDef{
			Name:             "log",
			ShortDescription: "Logs any value passed through the configured logger",
			Icon:             "align-center",
			Tags:             []string{"system"},
			DocURL:           "https://bitspark.de/slang/docs/operator/log",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type:    "generic",
					Generic: "valueType",
				},
				Out: core.TypeDef{
					Type:    "generic",
					Generic: "valueType",
				},
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			i := in.Pull()
			log.Print(i)
			out.Push(i)
		}
	},
}
