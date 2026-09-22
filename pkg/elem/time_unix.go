package elem

import (
	"time"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var timeUNIXMillisCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: uuid.MustParse("d58b458e-8b3a-49f3-a6e9-45e737354937"),
		Meta: core.BlueprintMetaDef{
			Name:             "UNIX milliseconds",
			ShortDescription: "emits the current UNIX timestamp in milliseconds",
			Icon:             "stamp",
			Tags:             []string{"time"},
			DocURL:           "https://bitspark.de/slang/docs/operator/unix-milliseconds",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "trigger",
				},
				Out: core.TypeDef{
					Type: "number",
				},
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			if i := in.Pull(); !core.IsMarker(i) {
				out.Push(float64(time.Now().UnixNano() / 1000 / 1000))
			} else {
				out.Push(i)
			}
		}
	},
}
