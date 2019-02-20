package elem

import (
	"context"
	"github.com/Bitspark/slang/pkg/core"
	_ "github.com/go-sql-driver/mysql"
	"github.com/lovoo/goka"
	"github.com/lovoo/goka/codec"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var databaseKafjaReadCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: "7aebae50-254d-4216-827a-96445be081da",
		Meta: core.OperatorMetaDef{
			Name: "Kafka subscribe",
			ShortDescription: "subscribes to a Kafka stream",
			Icon: "database",
			Tags: []string{"database", "kafka"},
			DocURL: "https://bitspark.de/slang/docs/operator/kafka-subscribe",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"trigger": {
							Type: "trigger",
						},
						"{queryParams}": {
							Type: "primitive",
						},
					},
				},
				Out: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "map",
						Map: map[string]*core.TypeDef{
							"trigger": {
								Type: "trigger",
							},
							"{rowColumns}": {
								Type: "primitive",
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
		topic := op.Property("topic").(goka.Stream)
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

			cb := func(ctx goka.Context, msg interface{}) {
				log.Printf("key = %s, msg = %v", ctx.Key(), msg)
				outKeyStream.Push(ctx.Key())
				outValueStream.Push(ctx.Value())
			}

			g := goka.DefineGroup("group",
				goka.Input(topic, new(codec.Bytes), cb),
				goka.Persist(new(codec.Int64)),
			)

			p, err := goka.NewProcessor(brokers, g)
			if err != nil {
				log.Fatalf("error creating processor: %v", err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan bool)

			go func() {
				defer close(done)
				if err = p.Run(ctx); err != nil {
					log.Fatalf("error running processor: %v", err)
				}
			}()

			wait := make(chan os.Signal, 1)
			signal.Notify(wait, syscall.SIGINT, syscall.SIGTERM)
			<-wait   // wait for SIGINT/SIGTERM
			cancel() // gracefully stop processor
			<-done

			out.Push(nil)
		}
	},
}
