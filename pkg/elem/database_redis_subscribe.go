package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/go-redis/redis"
)

var databaseRedisSubscribeCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: "eb3fd302-f6b0-4c2a-b353-ff0a01e49d09",
		Meta: core.OperatorMetaDef{
			Name:             "Redis Subscribe",
			ShortDescription: "executes an subscribe command at the specified Redis server",
			Icon:             "database",
			Tags:             []string{"database", "redis"},
			DocURL:           "https://bitspark.de/slang/docs/operator/redis-subscribe",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "trigger",
				},
				Out: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "string",
					},
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
			"channel": {
				Type: "string", // should be a map in the future as you can subscribe to multiple channels
			},
		},
	},
	opFunc: func(op *core.Operator) {
		host := op.Property("host").(string)
		password := op.Property("password").(string)
		channel := op.Property("channel").(string)

		// probably should use the clustered client
		client := redis.NewClient(&redis.Options{
			Addr:     host,     // localhost:6379
			Password: password, //
			DB:       0,        // use default DB
		})

		in := op.Main().In()
		out := op.Main().Out()

		pubsub := client.Subscribe(channel)
		ch := pubsub.Channel()
		defer pubsub.Close()

		// this loop can never break or end
		// as we are constantly wait for messages
		// on the subscribed channel(s)
		for {
			marker := in.Pull()
			if marker != nil {
				out.Push(marker)
				continue
			}
			out.PushBOS()
			for {
				select {
				case msg := <-ch:
					out.Stream().Push(msg.Payload)
				}
			}
			// it also makes no sense to push an EOS as
			// we never finish waiting for messages.
		}
	},
}
