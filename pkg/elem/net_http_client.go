package elem

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/Bitspark/slang/pkg/core"

	"github.com/google/uuid"
)

// netHTTPClientTimeout bounds one exchange, including reading the body. It stays
// below the hosted runtime's 15-second invocation limit, so a program receives a
// failed request instead of the runtime stopping its instance.
const netHTTPClientTimeout = 10 * time.Second

// netHTTPClientTimeoutOverride replaces the timeout when positive. Tests use it
// to observe a timeout without waiting for the default.
var netHTTPClientTimeoutOverride int64

func netHTTPTimeout() time.Duration {
	if override := atomic.LoadInt64(&netHTTPClientTimeoutOverride); override > 0 {
		return time.Duration(override)
	}
	return netHTTPClientTimeout
}

// netHTTPExchange sends one request and reads the whole response before the
// timeout expires or ctx is cancelled.
func netHTTPExchange(ctx context.Context, method, url string, body []byte, headers []interface{}) (*http.Response, []byte, error) {
	ctx, cancel := context.WithTimeout(ctx, netHTTPTimeout())
	defer cancel()

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

	resp, err := http.DefaultClient.Do(r)
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
	opFunc: func(op *core.Operator) {
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

			resp, respBody, err := netHTTPExchange(running, method, url, body, headers)
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
	},
}
