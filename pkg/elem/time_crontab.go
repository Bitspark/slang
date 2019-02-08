package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/robfig/cron"
)

var timeCrontabCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Meta: core.OperatorMetaDef{
			Name: "crontab",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "string",
				},
				Out: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "generic",
						Generic: "itemType",
					},
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{
			"handler": {
				Out: core.TypeDef{
					Type: "trigger",
				},
				In: core.TypeDef{
					Type: "generic",
					Generic: "itemType",
				},
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		c := cron.New()
		handler := op.Delegate("handler")
		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			crontab := i.(string)

			out.PushBOS()
			c.AddFunc(crontab, func() {
				handler.Out().Push(nil)

				item := handler.In().Pull()
				out.Stream().Push(item)
			})
			c.Start()

			op.WaitForStop()
			c.Stop()
			out.PushEOS()
			break
		}
	},
	opConnFunc: func(op *core.Operator, dst, src *core.Port) error {
		return nil
	},
}