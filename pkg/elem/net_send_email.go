package elem

import (
	"github.com/Bitspark/slang/pkg/core"
	"fmt"
	"net/smtp"
	"net/mail"
	"net"
	"crypto/tls"
	"github.com/Bitspark/slang/pkg/utils"
)

var netSendEmailCfg = &builtinConfig{
	opDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"to": {
							Type: "map",
							Map: map[string]*core.TypeDef{
								"name": {
									Type: "string",
								},
								"address": {
									Type: "string",
								},
							},
						},
						"from": {
							Type: "map",
							Map: map[string]*core.TypeDef{
								"name": {
									Type: "string",
								},
								"address": {
									Type: "string",
								},
							},
						},
						"subject": {
							Type: "string",
						},
						"body": {
							Type: "binary",
						},
					},
				},
				Out: core.TypeDef{
					Type: "boolean",
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
		PropertyDefs: core.TypeDefMap{
			"server": {
				Type: "string",
			},
			"username": {
				Type: "string",
			},
			"password": {
				Type: "string",
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		server := op.Property("server").(string)
		username := op.Property("username").(string)
		password := op.Property("password").(string)
		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			im := i.(map[string]interface{})

			fromMap := im["from"].(map[string]interface{})
			toMap := im["to"].(map[string]interface{})
			from := mail.Address{Name: fromMap["name"].(string), Address: fromMap["address"].(string)}
			to   := mail.Address{Name: toMap["name"].(string), Address: toMap["address"].(string)}
			subj := im["subject"].(string)
			body := im["body"].(utils.Binary)

			// Setup headers
			headers := make(map[string]string)
			headers["From"] = from.String()
			headers["To"] = to.String()
			headers["Subject"] = subj

			// Setup message
			message := ""
			for k,v := range headers {
				message += fmt.Sprintf("%s: %s\r\n", k, v)
			}
			message += "\r\n" + string(body)

			host, _, _ := net.SplitHostPort(server)

			auth := smtp.PlainAuth("",username, password, host)

			// TLS config
			tlsconfig := &tls.Config {
				InsecureSkipVerify: true,
				ServerName: host,
			}

			// Here is the key, you need to call tls.Dial instead of smtp.Dial
			// for smtp servers running on 465 that require an ssl connection
			// from the very beginning (no starttls)
			conn, err := tls.Dial("tcp", server, tlsconfig)
			if err != nil {
				out.Push(false)
				continue
			}

			c, err := smtp.NewClient(conn, host)
			if err != nil {
				out.Push(false)
				continue
			}

			// Auth
			if err = c.Auth(auth); err != nil {
				out.Push(false)
				continue
			}

			// To && From
			if err = c.Mail(from.Address); err != nil {
				out.Push(false)
				continue
			}

			if err = c.Rcpt(to.Address); err != nil {
				out.Push(false)
				continue
			}

			// Data
			w, err := c.Data()
			if err != nil {
				out.Push(false)
				continue
			}

			_, err = w.Write([]byte(message))
			if err != nil {
				out.Push(false)
				continue
			}

			err = w.Close()
			if err != nil {
				out.Push(false)
				continue
			}

			c.Quit()

			out.Push(true)
		}
	},
}
