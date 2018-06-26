package api

import "github.com/Bitspark/slang/pkg/core"

// This file contains functions and structs needed for constructing operator execution surroundings for http endpoints

// HttpRequest Slang type
var HTTP_REQUEST_DEF = core.TypeDef{
	Type: "map",
	Map: map[string]*core.TypeDef{
		"protocol": {
			Type: "string",
		},
		"method": {
			Type: "string",
		},
		"path": {
			Type: "string",
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
}

// HttpResponse Slang type
var HTTP_RESPONSE_DEF = core.TypeDef{
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
}

// Constructs an executable operator
// TODO: Make safer (maybe require an API key?)
func ConstructHttpEndpoint(env *Environ, port int, operator string, gens core.Generics, props core.Properties) (*core.OperatorDef, error) {
	httpDef := &core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "trigger",
				},
				Out: core.TypeDef{
					Type: "trigger",
				},
			},
		},
		Connections: make(map[string][]string),
	}

	path, err := env.GetOperatorPath(operator, "")
	if err != nil {
		return nil, err
	}

	// Build operator to get interface and see if it is free from errors
	// It will be compiled a second time later
	op, err := env.BuildAndCompileOperator(path, gens, props)
	if err != nil {
		return nil, err
	}

	// Const port instance
	portIns := &core.InstanceDef{
		Name: "port",
		Operator: "slang.const",
		Generics: core.Generics{
			"valueType": {
				Type: "number",
			},
		},
		Properties: core.Properties{
			"value": float64(port),
		},
	}
	httpDef.InstanceDefs = append(httpDef.InstanceDefs, portIns)
	httpDef.Connections["("] = []string{"(port"}

	// HTTP operator instance
	httpIns := &core.InstanceDef{
		Name: "httpServer",
		Operator: "slang.net.httpServer",
	}
	httpDef.InstanceDefs = append(httpDef.InstanceDefs, httpIns)
	httpDef.Connections["port)"] = []string{"(httpServer"}
	httpDef.Connections["httpServer)"] = []string{")"}

	// The HTTP server is connected now, only the handler delegate is missing

	// This is the actual operator we want to execute
	operatorIns := &core.InstanceDef{
		Name: "operator",
		Operator: operator,
		Generics: gens,
		Properties: props,
	}
	httpDef.InstanceDefs = append(httpDef.InstanceDefs, operatorIns)

	// Get operator interface
	inDef := op.Main().In().Define()
	outDef := op.Main().Out().Define()

	if inDef.Equals(HTTP_REQUEST_DEF) {
		// If the operator can handle HTTP requests itself, just pass them
		httpDef.Connections["httpServer.handler)~"] = []string{"(operator"}
	} else {
		// In this case we are not interested in anything but the body
		// It contains the JSON we need to unpack
		unpackerIns := &core.InstanceDef{
			Name: "unpacker",
			Operator: "slang.encoding.json.read",
			Generics: core.Generics{
				"itemType": &inDef,
			},
		}
		httpDef.InstanceDefs = append(httpDef.InstanceDefs, unpackerIns)
		httpDef.Connections["httpServer.handler)~.body"] = []string{"(unpacker"}
		httpDef.Connections["unpacker)"] = []string{"(operator"}
	}

	if outDef.Equals(HTTP_RESPONSE_DEF) {
		// If the operator produces HTTP responses itself, just pass them
		httpDef.Connections["operator)"] = []string{"~(httpServer.handler"}
	} else {
		// In this case we are not interested in anything but the body
		// It contains the JSON we need to pack
		unpackerIns := &core.InstanceDef{
			Name: "packer",
			Operator: "slang.encoding.json.write",
			Generics: core.Generics{
				"itemType": &outDef,
			},
		}
		httpDef.InstanceDefs = append(httpDef.InstanceDefs, unpackerIns)
		httpDef.Connections["operator)"] = []string{"(packer"}
		// We connect unpacker output later

		// Now we still need status (200) and default headers ([])

		// Status code operator
		statusCodeIns := &core.InstanceDef{
			Name: "statusCode",
			Operator: "slang.const",
			Generics: core.Generics{
				"valueType": {
					Type: "number",
				},
			},
			Properties: core.Properties{
				"value": 200,
			},
		}
		httpDef.InstanceDefs = append(httpDef.InstanceDefs, statusCodeIns)
		// We connect it later

		// Status code operator
		headersIns := &core.InstanceDef{
			Name: "headers",
			Operator: "slang.const",
			Generics: core.Generics{
				"valueType": {
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
			},
			Properties: core.Properties{
				"value": []interface{}{},
			},
		}
		httpDef.InstanceDefs = append(httpDef.InstanceDefs, headersIns)
		// We connect it later

		httpDef.Connections["packer)"] = []string{"~.body(httpServer.handler", "(statusCode", "(headers"}
		httpDef.Connections["statusCode)"] = []string{"~.status(httpServer.handler"}
		httpDef.Connections["headers)"] = []string{"~.headers(httpServer.handler"}
	}

	return httpDef, nil
}