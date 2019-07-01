package elem

import (
	"time"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

var timeDelayId = uuid.MustParse("7d61b83a-9aa2-4875-9c21-1e11f6adbfae")
var timeDelayCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: timeDelayId,
		Meta: core.OperatorMetaDef{
			Name:             "delay",
			ShortDescription: "takes an item and emits it again after a given number of milliseconds has passed",
			Icon:             "clock",
			Tags:             []string{"time"},
			DocURL:           "https://bitspark.de/slang/docs/operator/now",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"item": {
							Type:    "generic",
							Generic: "itemType",
						},
						"delay": {
							Type: "number",
						},
					},
				},
				Out: core.TypeDef{
					Type:    "generic",
					Generic: "itemType",
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

			im := i.(map[string]interface{})
			delay := im["delay"].(float64)
			item := im["item"]

			<-time.After(time.Millisecond * time.Duration(delay))
			out.Push(item)
		}
	},
	opConnFunc: func(op *core.Operator, dst, src *core.Port) error {
		return nil
	},
}
