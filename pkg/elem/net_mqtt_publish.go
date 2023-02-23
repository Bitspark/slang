package elem

import (
	"crypto/tls"

	"github.com/Bitspark/slang/pkg/core"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

var netMQTTPublishCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: uuid.MustParse("c6b5bef6-e93e-4bc1-8ded-49c90919f39d"),
		Meta: core.BlueprintMetaDef{
			Name:             "MQTT publish",
			ShortDescription: "publishes an MQTT message at a given topic",
			Icon:             "chart-network",
			Tags:             []string{"network"},
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
		PropertyDefs: core.PropertyMap{
			"broker": {
				TypeDef: core.TypeDef{
					Type: "string",
				},
			},
			"username": {
				TypeDef: core.TypeDef{
					Type: "string",
				},
			},
			"password": {
				TypeDef: core.TypeDef{
					Type: "string",
				},
			},
			"verifyCertificate": {
				TypeDef: core.TypeDef{
					Type:     "boolean",
					Optional: true,
				},
			},
			"clientCertificate": {
				TypeDef: core.TypeDef{
					Type:     "string",
					Optional: true,
				},
			},
			"clientKey": {
				TypeDef: core.TypeDef{
					Type:     "string",
					Optional: true,
				},
			},
			"caCertificate": {
				TypeDef: core.TypeDef{
					Type:     "string",
					Optional: true,
				},
			},
			// "clientId": {
			// 	Type: "string",
			// },
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
