package elem

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/Bitspark/slang/pkg/core"

	"github.com/google/uuid"
)

// netHTTPExchange sends one request through client and reads the whole response,
// within the client's timeout and until ctx is cancelled.
func netHTTPExchange(ctx context.Context, client *http.Client, method, url string, body []byte, headers []interface{}) (*http.Response, []byte, error) {
	r, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	for _, header := range headers {
		entry := header.(map[string]interface{})
		for _, value := range entry["values"].([]interface{}) {
			r.Header.Set(entry["key"].(string), value.(string))
		}
	}

	resp, err := client.Do(r)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	return resp, respBody, nil
}

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
	makeFunc: func(caps Capabilities) (core.OFunc, error) {
		if caps.HTTP == nil {
			return nil, errors.New("this host does not provide HTTP")
		}
		client := caps.HTTP
		return func(op *core.Operator) {
			in := op.Main().In()
			out := op.Main().Out()

			// Stopping the operator cancels the request it is waiting on.
			running, cancel := context.WithCancel(context.Background())
			defer cancel()
			go func() {
				select {
				case <-op.Done():
					cancel()
				case <-running.Done():
				}
			}()

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

				resp, respBody, err := netHTTPExchange(running, client, method, url, body, headers)
				if err != nil {
					if running.Err() == nil {
						op.Logger().Error(err)
					}
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
		}, nil
	},
}
