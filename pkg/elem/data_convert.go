package elem

import (
	"fmt"
	"github.com/Bitspark/slang/pkg/core"
	"strconv"
	"strings"
)

var dataConvertCfg = &builtinConfig{
	opDef: core.OperatorDef{
		Id: "d1191456-3583-4eaf-8ec1-e486c3818c60",
		Meta: core.OperatorMetaDef{
			Name: "convert",
			ShortDescription: "converts the type of a value",
			Icon: "arrow-alt-right",
			Tags: []string{"data"},
			DocURL: "https://bitspark.de/slang/docs/operator/convert",
		},
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
	opFunc: func(op *core.Operator) {
		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) || i == nil {
				out.Push(i)
				continue
			}

			switch in.Type() {
			case core.TYPE_NUMBER:
				item := i.(float64)
				switch out.Type() {
				case core.TYPE_STRING:
					out.Push(strconv.FormatFloat(item, 'f', -1, 64))
				default:
					panic("not supported yet")
				}
			case core.TYPE_BOOLEAN:
				item := i.(bool)
				switch out.Type() {
				case core.TYPE_STRING:
					out.Push(fmt.Sprintf("%v", item))
				case core.TYPE_BINARY:
					out.Push(core.Binary(fmt.Sprintf("%v", item)))
				default:
					panic("not supported yet")
				}
			case core.TYPE_STRING:
				item := i.(string)
				switch out.Type() {
				case core.TYPE_STRING:
					out.Push(item)
				case core.TYPE_BINARY:
					out.Push(core.Binary(item))
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
				item := i.(core.Binary)
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
					out.Push(core.Binary(fmt.Sprintf("%v", item)))
				default:
					panic("not supported yet")
				}
			default:
				panic("not supported yet")
			}
		}
	},
}
