package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"time"
)

func parseDate(dateStr string) (time.Time, error) {
	var err error
	var t time.Time
	for _, layout := range []string{time.ANSIC, time.UnixDate, time.RubyDate,
		time.RFC822, time.RFC822Z, time.RFC850, time.RFC1123, time.RFC1123Z, time.RFC3339, time.RFC3339Nano,
		time.Kitchen, time.Stamp, time.StampMilli, time.StampMicro, time.StampNano,
	} {
		t, err = time.Parse(layout, dateStr)
		if err == nil {
			return t, nil
		}
	}
	return t, err
}

var timeParseDateCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: "2a9da2d5-2684-4d2f-8a37-9560d0f2de29",
		Meta: core.OperatorMetaDef{
			Name: "to date",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "string",
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
				t, _ := parseDate(i.(string))
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
