package elem

import (
	"crypto/tls"

	"github.com/Bitspark/slang/pkg/core"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

var netMQTTSubscribeCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: uuid.MustParse("fd51e295-3483-4558-9b26-8c16d579c4ef"),
		Meta: core.BlueprintMetaDef{
			Name:             "MQTT subscribe",
			ShortDescription: "subscribes at a given topic, behaves like an MQTT client",
			Icon:             "chart-network",
			Tags:             []string{"network"},
			DocURL:           "https://bitspark.de/slang/docs/operator/mqtt-subscribe",
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
				},
			},
		},
		PropertyDefs: core.PropertyMap{
			"broker": {
				core.TypeDef{
					Type: "string",
				},
				nil,
			},
			"username": {
				core.TypeDef{
					Type: "string",
				},
				nil,
			},
			"password": {
				core.TypeDef{
					Type: "string",
				},
				nil,
			},
			"topic": {
				core.TypeDef{
					Type: "string",
				},
				nil,
			},
			"verifyCertificate": {
				core.TypeDef{
					Type:     "boolean",
					Optional: true,
				},
				nil,
			},
			"clientCertificate": {
				core.TypeDef{
					Type:     "string",
					Optional: true,
				},
				nil,
			},
			"clientKey": {
				core.TypeDef{
					Type:     "string",
					Optional: true,
				},
				nil,
			},
			"caCertificate": {
				core.TypeDef{
					Type:     "string",
					Optional: true,
				},
				nil,
			},
		},
	},
	opFunc: func(op *core.Operator) {
		options := mqtt.NewClientOptions().
			AddBroker(op.Property("broker").(string)).
			SetUsername(op.Property("username").(string)).
			SetPassword(op.Property("password").(string)).
			SetTLSConfig(&tls.Config{
				ClientAuth:         tls.NoClientCert,
				InsecureSkipVerify: true,
			})

		topic := op.Property("topic").(string)

		client := mqtt.NewClient(options)
		token := client.Connect().(*mqtt.ConnectToken)
		token.Wait()
		if token.Error() != nil {
			panic(token.Error())
		}

		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			in.Pull()

			out.PushBOS()
			outStream := out.Stream()
			client.Subscribe(topic, 2, func(client mqtt.Client, message mqtt.Message) {
				outStream.Map("messageId").Push(float64(message.MessageID()))
				outStream.Map("payload").Push(core.Binary(message.Payload()))
				outStream.Map("topic").Push(message.Topic())
			})

			op.WaitForStop()
			out.PushEOS()
			break
		}
	},
}
