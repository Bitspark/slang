package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/Bitspark/slang/pkg/utils"
)

var netMQTTSubscribeLegacyCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "trigger",
				},
				Out: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "generic",
						Generic: "itemType",
					},
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{
			"handler": {
				Out: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"messageId": {
							Type: "number",
						},
						"payload": {
							Type: "binary",
						},
						"topic": {
							Type: "string",
						},
					},
				},
				In: core.TypeDef{
					Type: "generic",
					Generic: "itemType",
				},
			},
		},
		PropertyDefs: map[string]*core.TypeDef{
			"broker": {
				Type: "string",
			},
			"username": {
				Type: "string",
			},
			"password": {
				Type: "string",
			},
			// "clientId": {
			// 	Type: "string",
			// },
			"topic": {
				Type: "string",
			},
		},
	},
	opFunc: func(op *core.Operator) {
		options := mqtt.NewClientOptions()
		options.AddBroker(op.Property("broker").(string))
		// options.SetClientID(op.Property("clientId").(string))
		options.SetUsername(op.Property("username").(string))
		options.SetPassword(op.Property("password").(string))

		topic := op.Property("topic").(string)

		client := mqtt.NewClient(options)
		token := client.Connect().(*mqtt.ConnectToken)
		token.Wait()
		if token.Error() != nil {
			panic(token.Error())
		}

		handler := op.Delegate("handler")

		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			in.Pull()
			out.PushBOS()

			client.Subscribe(topic, 2, func(client mqtt.Client, message mqtt.Message) {
				handler.Out().Map("messageId").Push(float64(message.MessageID()))
				handler.Out().Map("payload").Push(utils.Binary(message.Payload()))
				handler.Out().Map("topic").Push(message.Topic())

				// Push out item produced from handler
				out.Push(handler.In().Pull())
			})

			// TODO: When send EOS? In case of error?
		}
	},
}
