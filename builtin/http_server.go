package builtin

import (
	"slang/core"
	"strconv"
	"net/http"
	"bytes"
	"time"
)

type requestHandler struct {
	hOut *core.Port
	hIn *core.Port
}

func (r *requestHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	hIn := r.hIn
	hOut := r.hOut

	// Push out all request information
	hOut.Map("method").Push(req.Method)
	hOut.Map("path").Push(req.URL.String())
	hOut.Map("protocol").Push(req.Proto)

	hOut.Map("headers").PushBOS()
	headersOut := hOut.Map("headers").Stream()
	for key, val := range req.Header {
		headersOut.Map("key").Push(key)
		headersOut.Map("value").Push(val)
	}
	hOut.Map("headers").PushEOS()

	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)
	hOut.Map("body").Push(buf.String())

	// Gather all response information
	statusCode := hIn.Map("status").PullInt()
	resp.WriteHeader(statusCode)

	headers := hIn.Map("headers").Pull().([]interface{})
	for _, entry := range headers {
		header := entry.(map[string]interface{})
		resp.Header().Set(header["key"].(string), header["value"].(string))
	}

	resp.Write([]byte(hIn.Map("body").PullString()))
}

var httpServerOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		In: core.PortDef{
			Type: "number",
		},
		Out: core.PortDef{
			Type: "string",
		},
		Delegates: map[string]*core.DelegateDef{
			"handler": {
				In: core.PortDef{
					Type: "stream",
					Stream: &core.PortDef{
						Type: "map",
						Map: map[string]*core.PortDef{
							"status": {
								Type: "number",
							},
							"headers": {
								Type: "stream",
								Stream: &core.PortDef{
									Type: "map",
									Map: map[string]*core.PortDef{
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
								Type: "string",
							},
						},
					},
				},
				Out: core.PortDef{
					Type: "stream",
					Stream: &core.PortDef{
						Type: "map",
						Map: map[string]*core.PortDef{
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
								Stream: &core.PortDef{
									Type: "map",
									Map: map[string]*core.PortDef{
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
								Type: "string",
							},
						},
					},
				},
			},
		},
	},
	oFunc: func(in, out *core.Port, dels map[string]*core.Delegate, store interface{}) {
		slangHandler := dels["handler"]
		hOut := slangHandler.Out().Stream()
		hIn := slangHandler.In().Stream()

		for true {
			port := in.PullInt()

			// Once we receive the port, we signal start of request processing by pushing a BOS
			slangHandler.Out().PushBOS()
			slangHandler.In().PullBOS()

			s := &http.Server{
				Addr:           ":" + strconv.Itoa(port),
				Handler:        &requestHandler{hIn: hIn, hOut: hOut},
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
	oPropFunc: func(o *core.Operator, props map[string]interface{}) error {
		return nil
	},
}
