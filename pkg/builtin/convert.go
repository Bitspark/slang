package builtin

import (
	"github.com/Bitspark/slang/pkg/core"
	"strconv"
	"fmt"
	"strings"
)

var convertOpCfg = &builtinConfig{
	oDef: core.OperatorDef{
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type:    "generic",
					Generic: "fromType",
				},
				Out: core.TypeDef{
					Type:    "generic",
					Generic: "toType",
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
	},
	oFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			if i == nil {
				switch out.Type() {
				case core.TYPE_STRING:
					out.Push("")
				case core.TYPE_BINARY:
					out.Push([]byte{})
				case core.TYPE_NUMBER:
					out.Push(0.0)
				default:
					panic("not supported yet")
				}
				continue
			}

			switch in.Type() {
			case core.TYPE_BOOLEAN:
				item := i.(bool)
				switch out.Type() {
				case core.TYPE_STRING:
					out.Push(fmt.Sprintf("%v", item))
				case core.TYPE_BINARY:
					out.Push([]byte(fmt.Sprintf("%v", item)))
				default:
					panic("not supported yet")
				}
			case core.TYPE_STRING:
				item := i.(string)
				switch out.Type() {
				case core.TYPE_STRING:
					out.Push(item)
				case core.TYPE_BINARY:
					out.Push([]byte(item))
				case core.TYPE_NUMBER:
					item = strings.Trim(item, " ")
					floatItem := 0.0
					if strings.Contains(item, ":") {
						items := strings.Split(item, ":")
						factor := 1.0
						for i := len(items) - 1; i >= 0; i-- {
							part, _ := strconv.ParseFloat(items[i], 64)
							floatItem += factor * part
							factor *= 60
						}
					} else {
						floatItem, _ = strconv.ParseFloat(item, 64)
					}
					out.Push(floatItem)
				default:
					panic("not supported yet")
				}
			case core.TYPE_BINARY:
				item := i.([]byte)
				switch out.Type() {
				case core.TYPE_STRING:
					out.Push(string(item))
				case core.TYPE_BINARY:
					out.Push(item)
				default:
					panic("not supported yet")
				}
			case core.TYPE_STREAM:
				item := i.([]interface{})
				switch out.Type() {
				case core.TYPE_STRING:
					out.Push(fmt.Sprintf("%v", item))
				case core.TYPE_BINARY:
					out.Push([]byte(fmt.Sprintf("%v", item)))
				default:
					panic("not supported yet")
				}
			default:
				panic("not supported yet")
			}
		}
	},
}