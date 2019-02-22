package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/go-redis/redis"
)

var databaseRedisHGetCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"key": {
							Type: "string",
						},
						"field": {
							Type: "string",
						},
					},
				},
				Out: core.TypeDef{
					Type: "string",
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{
			"creator": {
				Out: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"key": {
							Type: "string",
						},
						"field": {
							Type: "string",
						},
					},
				},
				In: core.TypeDef{
					Type: "string",
				},
			},
		},
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

			key := pair["key"].(string)
			field := pair["field"].(string)

			valueCmd := client.HGet(key, field)
			if valueCmd == nil {
				op.Delegate("creator").Out().Push(pair)
				value := op.Delegate("creator").In().Pull()

				boolCmd := client.HSet(key, field, value)

				if _, err := boolCmd.Result(); err == nil {
					out.Push(value)
				} else {
					panic(err)
				}
			} else {
				out.Push(valueCmd.String())
			}
		}
	},
}
