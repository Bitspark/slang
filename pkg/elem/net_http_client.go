package elem

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/log"
	"github.com/google/uuid"
)

var netHTTPClientCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: uuid.MustParse("f7f5907d-758b-4892-8a3e-ae86b877b869"),
		Meta: core.BlueprintMetaDef{
			Name:             "HTTP client",
			ShortDescription: "sends an HTTP request",
			Icon:             "browser",
			Tags:             []string{"network"},
			DocURL:           "https://bitspark.de/slang/docs/operator/http-client",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: func() core.TypeDef {
					req := HTTP_REQUEST_DEF.Copy()
					delete(req.Map, "params")
					delete(req.Map, "path")
					delete(req.Map, "query")
					req.Map["url"] = &core.TypeDef{Type: "string"}
					return req
				}(),
				Out: HTTP_RESPONSE_DEF.Copy(),
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			req := i.(map[string]interface{})
			method := req["method"].(string)
			url := req["url"].(string)
			body := req["body"].(core.Binary)
			headers := req["headers"].([]interface{})

			r, err := http.NewRequest(method, url, bytes.NewReader(body))
			if err != nil {
				log.Error(err)
				out.Push(nil)
				continue
			}
			for _, header := range headers {
				entry := header.(map[string]interface{})
				for _, value := range entry["values"].([]interface{}) {
					r.Header.Set(entry["key"].(string), value.(string))
				}
			}

			resp, err := http.DefaultClient.Do(r)
			
			if err != nil {
				log.Error(err)
				out.Push(nil)
				continue
			}

			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Error(err)
				out.Push(nil)
				continue
			}

			out.Map("status").Push(float64(resp.StatusCode))
			out.Map("body").Push(core.Binary(respBody))

			out.Map("headers").PushBOS()
			for key := range resp.Header {
				out.Map("headers").Stream().Map("key").Push(key)
				out.Map("headers").Stream().Map("value").Push(resp.Header.Get(key))
			}
			out.Map("headers").PushEOS()
		}
	},
}
