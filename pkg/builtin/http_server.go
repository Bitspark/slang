package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
	"strconv"
	"net/http"
	"bytes"
	"time"
)

type requestHandler struct {
	hOut *core.Port
	hIn  *core.Port
	sync *core.Synchronizer
}

func (r *requestHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	req.ParseForm()

	token := r.sync.Push(func(out *core.Port) {
		// Push out all request information
		out.Map("method").Push(req.Method)
		out.Map("path").Push(req.URL.Path)
		out.Map("protocol").Push(req.Proto)

		out.Map("headers").PushBOS()
		headersOut := out.Map("headers").Stream()
		for key, val := range req.Header {
			headersOut.Map("key").Push(key)
			headersOut.Map("value").Push(val)
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
		out.Map("body").Push(buf.Bytes())
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

var httpServerOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
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
				In: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "map",
						Map: map[string]*core.TypeDef{
							"status": {
								Type: "number",
							},
							"headers": {
								Type: "stream",
								Stream: &core.TypeDef{
									Type: "map",
									Map: map[string]*core.TypeDef{
										"key": {
											Type: "string",
										},
										"value": {
											Type: "string",
										},
									},
								},
							},
							"body": {
								Type: "binary",
							},
						},
					},
				},
				Out: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "map",
						Map: map[string]*core.TypeDef{
							"method": {
								Type: "string",
							},
							"path": {
								Type: "string",
							},
							"protocol": {
								Type: "string",
							},
							"headers": {
								Type: "stream",
								Stream: &core.TypeDef{
									Type: "map",
									Map: map[string]*core.TypeDef{
										"key": {
											Type: "string",
										},
										"value": {
											Type: "string",
										},
									},
								},
							},
							"body": {
								Type: "binary",
							},
							"params": {
								Type: "stream",
								Stream: &core.TypeDef{
									Type: "map",
									Map: map[string]*core.TypeDef{
										"key": {
											Type: "string",
										},
										"values": {
											Type: "stream",
											Stream: &core.TypeDef{
												Type: "string",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	},
	oFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		slangHandler := op.Delegate("handler")
		sync := &core.Synchronizer{}
		sync.Init(
			slangHandler.In().Stream(),
			slangHandler.Out().Stream())
		go sync.Worker()

		for {
			port, marker := in.PullInt()
			if marker != nil {
				out.Push(marker)
				continue
			}

			// Once we receive the port, we signal start of request processing by pushing a BOS
			slangHandler.Out().PushBOS()
			slangHandler.In().PullBOS()

			s := &http.Server{
				Addr:           ":" + strconv.Itoa(port),
				Handler:        &requestHandler{sync: sync},
				ReadTimeout:    10 * time.Second,
				WriteTimeout:   10 * time.Second,
				MaxHeaderBytes: 1 << 20,
			}
			err := s.ListenAndServe()
			out.Push(err.Error())

			// Once we terminate, we signal end of request processing by pushing an EOS
			slangHandler.Out().PushEOS()
			slangHandler.In().PullEOS()
		}
	},
}
