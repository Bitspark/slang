package elem

import (
	"time"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var timeDateNowCfg = &builtinConfig{
	safe: true,
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
				Out: timeParseDateCfg.blueprint.ServiceDefs[core.MAIN_SERVICE].Out,
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		odate := out.Map("date")
		otime := out.Map("time")
		for !op.CheckStop() {
			if i := in.Pull(); !core.IsMarker(i) {
				t := time.Now()
				odate.Map("year").Push(t.Year())
				odate.Map("month").Push(int(t.Month()))
				odate.Map("day").Push(t.Day())
				otime.Map("hour").Push(t.Hour())
				otime.Map("minute").Push(t.Minute())
				otime.Map("second").Push(t.Second())
				out.Map("weekday").Push(t.Weekday().String())
			} else {
				out.Push(i)
			}
		}
	},
}
