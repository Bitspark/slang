package elem

import "github.com/Bitspark/slang/pkg/core"

// HttpRequest Slang type
var HTTP_REQUEST_DEF = core.TypeDef{
	Type: "map",
	Map: map[string]*core.TypeDef{
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
					"values": {
						Type: "stream",
						Stream: &core.TypeDef{
							Type: "string",
						},
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
