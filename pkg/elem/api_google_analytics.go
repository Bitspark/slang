package elem

import (
	"context"
	"github.com/Bitspark/slang/pkg/core"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/analytics/v3"
	"io/ioutil"
	"log"
	"strconv"
)

var apiGoogleAnalyticsCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: "257b55e8-b27f-4ef2-8581-5ef3f0af4491",
		Meta: core.OperatorMetaDef{
			Name: "visitors today",
			ShortDescription: "",
			Icon: "",
			Tags: []string{""},
			DocURL: "",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "trigger",
				},
				Out: core.TypeDef{
					Type: "number",
				},
			},
		},
		PropertyDefs: map[string]*core.TypeDef{
			"jsonFile": {
				Type: "string",
			},
			"gaid": {
				Type: "string",
			},
		},
	},
	opFunc: func(op *core.Operator) {
		data, err := ioutil.ReadFile(op.Property("jsonFile").(string))
		if err != nil {
			log.Fatal(err)
		}
		conf, err := google.JWTConfigFromJSON(data, "https://www.googleapis.com/auth/analytics.readonly")
		if err != nil {
			log.Fatal(err)
		}
		ctx := context.Background()
		client := conf.Client(ctx)

		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			svc, err := analytics.New(client)
			if err != nil {
				panic(err)
			}
			val := svc.Data.Ga.Get(op.Property("gaid").(string), "2019-03-12", "2019-03-13", "ga:sessions")
			rlt, err := val.Do()
			if err != nil {
				panic(err)
			}

			users, ok := rlt.TotalsForAllResults["ga:sessions"]
			if ok {
				userNumber, _ := strconv.Atoi(users)
				out.Push(float64(userNumber))
			} else {
				out.Push(nil)
			}
		}
	},
}
