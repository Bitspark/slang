package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

var databaseRedisHIncrByCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: uuid.MustParse("8d9e4c6e-20a2-44b1-8d51-ed98f4d3b4d8"),
		Meta: core.OperatorMetaDef{
			Name:             "Redis HIncr",
			ShortDescription: "executes an HIncr command at the specified Redis server",
			Icon:             "database",
			Tags:             []string{"database", "redis"},
			DocURL:           "https://bitspark.de/slang/docs/operator/redis-hincrby",
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
							Type: "number",
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
			intCmd := client.HIncrBy(pair["key"].(string), pair["field"].(string), int64(pair["value"].(float64)))

			if rlt, err := intCmd.Result(); err == nil {
				out.Push(float64(rlt))
			} else {
				panic(err)
			}
		}
	},
}
