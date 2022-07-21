package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

var databaseRedisGetCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: uuid.MustParse("362482c1-2021-4e5c-9463-b580a6c1967e"),
		Meta: core.BlueprintMetaDef{
			Name:             "Redis Get",
			ShortDescription: "executes a Get command at the specified Redis server",
			Icon:             "database",
			Tags:             []string{"database", "redis"},
			DocURL:           "https://bitspark.de/slang/docs/operator/redis-get",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "string",
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

			key := i.(string)
			if value, err := client.Get(key).Result(); err == nil {
				out.Push(value)
			} else {
				panic(err)
			}
		}
	},
}
