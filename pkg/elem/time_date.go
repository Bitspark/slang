package elem

import (
	"time"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
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
	safe: true,
	blueprint: core.Blueprint{
		Id: uuid.MustParse("2a9da2d5-2684-4d2f-8a37-9560d0f2de29"),
		Meta: core.BlueprintMetaDef{
			Name:             "to date",
			ShortDescription: "takes a string containing date and time and emits its parsed values",
			Icon:             "calendar-week",
			Tags:             []string{"time"},
			DocURL:           "https://bitspark.de/slang/docs/operator/to-date",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "string",
				},
				Out: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"weekday": {Type: "string"},
						"date": {
							Type: "map",
							Map: core.TypeDefMap{
								"year":  {Type: "number"},
								"month": {Type: "number"},
								"day":   {Type: "number"},
							},
						},
						"time": {
							Type: "map",
							Map: core.TypeDefMap{
								"hour":   {Type: "number"},
								"minute": {Type: "number"},
								"second": {Type: "number"},
							},
						},
					},
				},
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
				t, _ := parseDate(i.(string))
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
