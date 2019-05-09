package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var netMQTTPublishCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: "c6b5bef6-e93e-4bc1-8ded-49c90919f39d",
		Meta: core.OperatorMetaDef{
			Name:             "MQTT publish",
			ShortDescription: "publishes an MQTT message at a given topic",
			Icon:             "chart-network",
			Tags:             []string{"network", "mqtt"},
			DocURL:           "https://bitspark.de/slang/docs/operator/mqtt-publish",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"topic": {
							Type: "string",
						},
						"payload": {
							Type: "binary",
						},
					},
				},
				Out: core.TypeDef{
					Type: "number",
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
		},
	},
	opFunc: func(op *core.Operator) {
		options := mqtt.NewClientOptions()
		options.AddBroker(op.Property("broker").(string))
		// options.SetClientID(op.Property("clientId").(string))
		options.SetUsername(op.Property("username").(string))
		options.SetPassword(op.Property("password").(string))

		client := mqtt.NewClient(options)
		token := client.Connect().(*mqtt.ConnectToken)
		token.Wait()
		if token.Error() != nil {
			panic(token.Error())
		}

		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			im := i.(map[string]interface{})
			topic := im["topic"].(string)
			payload := im["payload"].(core.Binary)

			token := client.Publish(topic, 2, false, []byte(payload)).(*mqtt.PublishToken)
			token.Wait()
			out.Push(float64(token.MessageID()))
		}
	},
}
