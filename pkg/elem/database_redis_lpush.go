package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/go-redis/redis"
)

var databaseRedisLPushCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"key": {
							Type: "string",
						},
						"value": {
							Type: "string",
						},
					},
				},
				Out: core.TypeDef{
					Type: "number",
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
		PropertyDefs: map[string]*core.TypeDef{
			"host": {
				Type: "string",
			},
			"password": {
				Type: "string",
			},
		},
	},
	opFunc: func(op *core.Operator) {
		host := op.Property("host").(string)
		password := op.Property("password").(string)

		client := redis.NewClient(&redis.Options{
			Addr:     host,     // localhost:6379
			Password: password, //
			DB:       0,        // use default DB
		})

		in := op.Main().In()
		out := op.Main().Out()
		for {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			pair := i.(map[string]interface{})
			intCmd := client.LPush(pair["key"].(string), pair["value"].(string))

			if rlt, err := intCmd.Result(); err == nil {
				out.Push(float64(rlt))
			} else {
				panic(err)
			}
		}
	},
}
