package elem

import (
	"bytes"
	"net/http"
	"strconv"
	"time"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

type requestHandler struct {
	sync *core.Synchronizer
}

func (r *requestHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	req.ParseForm()

	// CORS
	if req.Method == "OPTIONS" {
		resp.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		resp.Header().Set("Access-Control-Allow-Methods", "*")
		resp.Header().Set("Access-Control-Allow-Origin", "*")
		resp.WriteHeader(200)
		resp.Write([]byte{})
		return
	}

	token := r.sync.Push(func(out *core.Port) {
		// Push out all request information
		out.Map("method").Push(req.Method)
		out.Map("path").Push(req.URL.Path)
		out.Map("query").Push(req.URL.RawQuery)

		out.Map("headers").PushBOS()
		headersOut := out.Map("headers").Stream()
		for key, vals := range req.Header {
			headersOut.Map("key").Push(key)
			headersOut.Map("values").PushBOS()
			valuesOut := headersOut.Map("values").Stream()
			for _, val := range vals {
				valuesOut.Push(val)
			}
			headersOut.Map("values").PushEOS()
		}
		out.Map("headers").PushEOS()

		out.Map("params").PushBOS()
		paramsOut := out.Map("params").Stream()
		for key, vals := range req.Form {
			paramsOut.Map("key").Push(key)
			paramsOut.Map("values").PushBOS()
			valuesOut := paramsOut.Map("values").Stream()
			for _, val := range vals {
				valuesOut.Push(val)
			}
			paramsOut.Map("values").PushEOS()
		}
		out.Map("params").PushEOS()

		buf := new(bytes.Buffer)
		buf.ReadFrom(req.Body)
		out.Map("body").Push(core.Binary(buf.Bytes()))
	})

	r.sync.Pull(token, func(in *core.Port) {
		// Gather all response information
		statusCode, _ := in.Map("status").PullInt()

		headers := in.Map("headers").Pull().([]interface{})
		for _, entry := range headers {
			header := entry.(map[string]interface{})
			resp.Header().Set(header["key"].(string), header["value"].(string))
		}

		resp.WriteHeader(statusCode)
		body, _ := in.Map("body").PullBinary()
		resp.Write([]byte(body))
	})
}

var netHTTPServerId = uuid.MustParse("241cc7ef-c6d6-49c1-8729-c5e3c0be8188")
var netHTTPServerCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: netHTTPServerId,
		Meta: core.BlueprintMetaDef{
			Name:             "HTTP server",
			ShortDescription: "starts an HTTP server, uses a handler delegate to process requests",
			Icon:             "server",
			Tags:             []string{"network", "http"},
			DocURL:           "https://bitspark.de/slang/docs/operator/http-server",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "number",
				},
				Out: core.TypeDef{
					Type: "string",
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{
			"handler": {
				In:  HTTP_RESPONSE_DEF.Copy(),
				Out: HTTP_REQUEST_DEF.Copy(),
			},
		},
	},
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		slangHandler := op.Delegate("handler")
		sync := &core.Synchronizer{}
		sync.Init(
			slangHandler.In(),
			slangHandler.Out())
		go sync.Worker()

		for !op.CheckStop() {
			port, marker := in.PullInt()
			if marker != nil {
				out.Push(marker)
				continue
			}

			s := &http.Server{
				Addr:           ":" + strconv.Itoa(port),
				Handler:        &requestHandler{sync: sync},
				ReadTimeout:    10 * time.Second,
				WriteTimeout:   10 * time.Second,
				MaxHeaderBytes: 1 << 20,
			}

			go func() {
				op.WaitForStop()
				s.Close()
			}()

			err := s.ListenAndServe()
			out.Push(err.Error())
		}
	},
}
