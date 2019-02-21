package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/utils"
	"github.com/Shopify/sarama"
	"log"
	"os"
	"os/signal"
)

var databaseKafjaSubscribeCfg = &builtinConfig{
	opDef: core.OperatorDef{
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
		PropertyDefs: map[string]*core.TypeDef{
			"brokers": {
				Type: "stream",
				Stream: &core.TypeDef{
					Type: "string",
				},
			},
			"topic": {
				Type: "string",
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

		for true {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			config := sarama.NewConfig()
			config.Net.LocalAddr = nil

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
					log.Printf("Consumed message offset %d\n", msg.Offset)

					outKeyStream.Push(msg.Key)
					outValueStream.Push(utils.Binary(msg.Value))
				case <-signals:
					break ConsumerLoop
				}
			}
			out.PushEOS()
		}
	},
}
