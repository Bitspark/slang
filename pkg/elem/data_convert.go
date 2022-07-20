package elem

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

func numberToBinary(value float64) core.Binary {
	bits := math.Float64bits(value)
	bytes := make(core.Binary, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	return bytes
}

func numberToString(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func stringToNumber(value string) float64 {
	result, _ := strconv.ParseFloat(value, 64)
	return result
}

func boolToNumber(value bool) float64 {
	if value {
		return 1.0
	} else {
		return 0.0
	}
}

func binaryToNumber(value core.Binary) float64 {
	bits := binary.LittleEndian.Uint64(value)
	return math.Float64frombits(bits)
}

func binaryToBool(value core.Binary) bool {
	return stringToBool(string(value))
}

func stringToBool(value string) bool {
	result, _ := strconv.ParseBool(value)
	return result
}

var dataConvertId = uuid.MustParse("d1191456-3583-4eaf-8ec1-e486c3818c60")
var dataConvertCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: dataConvertId,
		Meta: core.BlueprintMetaDef{
			Name:             "convert",
			ShortDescription: "converts the type of a value",
			Icon:             "arrow-alt-right",
			Tags:             []string{"data"},
			DocURL:           "https://bitspark.de/slang/docs/operator/convert",
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

			if in.Type() == out.Type() {
				out.Push(i)
				continue
			}

			switch in.Type() {
			case core.TYPE_NUMBER:
				value := i.(float64)
				switch out.Type() {
				case core.TYPE_STRING: // number -> string
					out.Push(numberToString(value))
				case core.TYPE_BINARY: // number -> binary
					out.Push(numberToBinary(value))
				case core.TYPE_BOOLEAN: // number -> bool
					out.Push(value != 0.0)
				default:
					panic("not supported yet")
				}
			case core.TYPE_BOOLEAN:
				value := i.(bool)
				switch out.Type() {
				case core.TYPE_STRING: // bool -> string
					out.Push(strconv.FormatBool(value))
				case core.TYPE_BINARY: // bool -> binary
					out.Push(core.Binary(strconv.FormatBool(value)))
				case core.TYPE_NUMBER: // bool -> number
					out.Push(boolToNumber(value))
				default:
					panic("not supported yet")
				}
			case core.TYPE_STRING:
				value := i.(string)
				switch out.Type() {
				case core.TYPE_BINARY: // string -> binary
					out.Push(core.Binary(value))
				case core.TYPE_NUMBER: // string -> number
					out.Push(stringToNumber(value))
				case core.TYPE_BOOLEAN: // string -> bool
					out.Push(stringToBool(value))
				default:
					panic("not supported yet")
				}
			case core.TYPE_BINARY:
				value := i.(core.Binary)
				switch out.Type() {
				case core.TYPE_STRING: // binary -> string
					out.Push(string(value))
				case core.TYPE_NUMBER: // binary -> number
					out.Push(binaryToNumber(value))
				case core.TYPE_BOOLEAN: // binary -> bool
					out.Push(binaryToBool(value))
				default:
					panic("not supported yet")
				}
			case core.TYPE_STREAM:
				value := i.([]interface{})
				switch out.Type() {
				case core.TYPE_STRING: // stream -> string
					out.Push(fmt.Sprintf("%v", value))
				case core.TYPE_BINARY: // stream -> binary
					out.Push(core.Binary(fmt.Sprintf("%v", value)))
				default:
					panic("not supported yet")
				}
			case core.TYPE_PRIMITIVE:
				switch out.Type() {
				case core.TYPE_STRING:
					switch value := i.(type) {
					case float64: // number -> string
						out.Push(numberToString(value))
					case string: // string -> string
						out.Push(value)
					case bool: // bool -> string
						out.Push(strconv.FormatBool(value))
					case core.Binary: // binary -> string
						out.Push(string(value))
					default:
						panic("not supported yet")
					}
				case core.TYPE_NUMBER:
					switch value := i.(type) {
					case float64: // number -> number
						out.Push(value)
					case string: // string -> number
						out.Push(stringToNumber(value))
					case bool: // bool -> number
						out.Push(boolToNumber(value))
					case core.Binary: // binary -> number
						out.Push(binaryToNumber(value))
					default:
						panic("not supported yet")
					}
				case core.TYPE_BOOLEAN:
					switch value := i.(type) {
					case float64: // number -> bool
						out.Push(value != 0.0)
					case string: // string -> bool
						out.Push(stringToBool(value))
					case bool: // bool -> bool
						out.Push(value)
					case core.Binary: // binary -> bool
						out.Push(binaryToBool(value))
					default:
						panic("not supported yet")
					}
				case core.TYPE_BINARY:
					switch value := i.(type) {
					case float64: // number -> binary
						out.Push(numberToBinary(value))
					case string: // string -> binary
						out.Push(core.Binary(value))
					case bool: // bool -> binary
						stringBool := strconv.FormatBool(value)
						out.Push(core.Binary(stringBool))
					case core.Binary: // binary -> binary
						out.Push(value)
					default:
						panic("not supported yet")
					}
				default:
					panic("not supported yet")
				}
			default:
				panic("not supported yet")
			}
		}
	},
}
