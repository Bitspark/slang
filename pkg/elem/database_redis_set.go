package elem

import (
	"time"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

var databaseRedisSetCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: uuid.MustParse("cdbf3e0d-1ce0-4565-9df6-d0e829c730e5"),
		Meta: core.OperatorMetaDef{
			Name:             "Redis Set",
			ShortDescription: "executes a Set command at the specified Redis server",
			Icon:             "database",
			Tags:             []string{"database", "redis"},
			DocURL:           "https://bitspark.de/slang/docs/operator/redis-set",
		},
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
					Type: "string",
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
			statusCmd := client.Set(pair["key"].(string), pair["value"].(string), 10*time.Second)

			if rlt, err := statusCmd.Result(); err == nil {
				out.Push(rlt)
			} else {
				panic(err)
			}
		}
	},
}
