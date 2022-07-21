package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
	"github.com/robfig/cron"
)

var timeCrontabCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: uuid.MustParse("60b849fd-ca5a-4206-8312-996e4e3f6c31"),
		Meta: core.BlueprintMetaDef{
			Name:             "crontab",
			ShortDescription: "takes a UNIX crontab string, sends triggers to its handler delegate accordingly",
			Icon:             "calendar-alt",
			Tags:             []string{"time"},
			DocURL:           "https://bitspark.de/slang/docs/operator/crontab",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "string",
				},
				Out: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type:    "generic",
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
					Type:    "generic",
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
