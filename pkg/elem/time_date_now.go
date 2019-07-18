package elem

import (
	"time"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var timeDateNowCfg = &builtinConfig{
	blueprint: core.Blueprint{
		Id: uuid.MustParse("808c7846-db9f-43ee-989b-37a08ce7e70d"),
		Meta: core.BlueprintMetaDef{
			Name:             "now",
			ShortDescription: "emits the current date and time",
			Icon:             "clock",
			Tags:             []string{"time"},
			DocURL:           "https://bitspark.de/slang/docs/operator/now",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "trigger",
				},
				Out: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"year":       {Type: "number"},
						"month":      {Type: "number"},
						"day":        {Type: "number"},
						"hour":       {Type: "number"},
						"minute":     {Type: "number"},
						"second":     {Type: "number"},
						"nanosecond": {Type: "number"},
					},
				},
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			if i := in.Pull(); !core.IsMarker(i) {
				t := time.Now()
				out.Map("year").Push(t.Year())
				out.Map("month").Push(int(t.Month()))
				out.Map("day").Push(t.Day())
				out.Map("hour").Push(t.Hour())
				out.Map("minute").Push(t.Minute())
				out.Map("second").Push(t.Second())
				out.Map("nanosecond").Push(t.Nanosecond())
			} else {
				out.Push(i)
			}
		}
	},
}
