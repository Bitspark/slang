package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

var databaseRedisHSetCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: uuid.MustParse("a6b45f70-e20c-40a5-ac39-c00068d10c81"),
		Meta: core.BlueprintMetaDef{
			Name:             "Redis HSet",
			ShortDescription: "executes an HSet command at the specified Redis server",
			Icon:             "database",
			Tags:             []string{"database", "redis"},
			DocURL:           "https://bitspark.de/slang/docs/operator/redis-hset",
		},
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
						"value": {
							Type: "string",
						},
					},
				},
				Out: core.TypeDef{
					Type: "boolean",
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
			boolCmd := client.HSet(pair["key"].(string), pair["field"].(string), pair["value"].(string))

			if rlt, err := boolCmd.Result(); err == nil {
				out.Push(rlt)
			} else {
				panic(err)
			}
		}
	},
}
