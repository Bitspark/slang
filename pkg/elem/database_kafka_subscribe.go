package elem

import (
	"os"
	"os/signal"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Shopify/sarama"
	"github.com/google/uuid"
)

var databaseKafkaSubscribeId = uuid.MustParse("b6cb78ca-bbfd-475e-a11f-3593ce295e3c")
var databaseKafkaSubscribeCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: databaseKafkaSubscribeId,
		Meta: core.BlueprintMetaDef{
			Name:             "Kafka subscribe",
			ShortDescription: "subscribes at a Kafka topic",
			Icon:             "",
			Tags:             []string{"database", "kafka"},
			DocURL:           "https://bitspark.de/slang/docs/operator/kafka-subscribe",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "trigger",
				},
				Out: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "map",
						Map: map[string]*core.TypeDef{
							"key": {
								Type: "string",
							},
							"value": {
								Type: "binary",
							},
						},
					},
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
		PropertyDefs: core.PropertyMap{
			"brokers": {
				TypeDef: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "string",
					},
				},
			},
			"topic": {
				TypeDef: core.TypeDef{
					Type: "string",
				},
			},
		},
	},
	opFunc: func(op *core.Operator) {
		topic := op.Property("topic").(string)
		brokers := []string{}

		for _, broker := range op.Property("brokers").([]interface{}) {
			brokers = append(brokers, broker.(string))
		}

		in := op.Main().In()
		out := op.Main().Out()

		outKeyStream := out.Stream().Map("key")
		outValueStream := out.Stream().Map("value")

		for {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			config := sarama.NewConfig()
			consumer, err := sarama.NewConsumer(brokers, config)
			if err != nil {
				panic(err)
			}

			partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
			if err != nil {
				panic(err)
			}

			signals := make(chan os.Signal, 1)
			signal.Notify(signals, os.Interrupt)

			out.PushBOS()
		ConsumerLoop:
			for {
				select {
				case msg := <-partitionConsumer.Messages():
					outKeyStream.Push(msg.Key)
					outValueStream.Push(core.Binary(msg.Value))
				case <-signals:
					break ConsumerLoop
				}
			}
			out.PushEOS()
		}
	},
}
