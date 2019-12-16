package elem

import (
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
)

func convertOperator(t *testing.T, from string, to string, in interface{}, out interface{}) {
	a := assertions.New(t)
	fo, _ := buildOperator(core.InstanceDef{
		Operator: dataConvertId,
		Generics: map[string]*core.TypeDef{
			"fromType": {
				Type: from,
			},
			"toType": {
				Type: to,
			},
		},
	})
	fo.Main().Out().Bufferize()
	go fo.Start()
	fo.Main().In().Push(in)

	a.PortPushes(out, fo.Main().Out())

}

func Test_Convert__String_To_Number(t *testing.T) {
	convertOperator(t, "string", "number", "2", 2.0)

	convertOperator(t, "string", "number", "asd", 0.0)
	convertOperator(t, "string", "number", "asd123", 0.0)
	convertOperator(t, "string", "number", "123asd123", 0.0)

	// This should work IMO
	convertOperator(t, "string", "number", "1.2-e2", 0.0)

	convertOperator(t, "string", "number", "1.2", 1.2)
	convertOperator(t, "string", "number", "1,231.2", 0.0)
}

func Test_Convert__To_Number(t *testing.T) {
	convertOperator(t, "primitive", "number", "2.1", 2.1)
	convertOperator(t, "string", "number", "2.1", 2.1)

	convertOperator(t, "primitive", "number", 2.1, 2.1)
	convertOperator(t, "number", "number", 2.1, 2.1)

	convertOperator(t, "primitive", "number", true, 1.0)
	convertOperator(t, "boolean", "number", true, 1.0)

	convertOperator(t, "primitive", "number", false, 0.0)
	convertOperator(t, "boolean", "number", false, 0.0)

	convertOperator(t, "binary", "number", core.Binary{0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40}, 3.141592653589793)
	convertOperator(t, "primitive", "number", nil, nil)
}

func Test_Convert_To_String(t *testing.T) {
	convertOperator(t, "primitive", "string", "2.1", "2.1")
	convertOperator(t, "string", "string", "2.1", "2.1")

	convertOperator(t, "primitive", "string", 2.1, "2.1")
	convertOperator(t, "number", "string", 2.1, "2.1")

	convertOperator(t, "primitive", "string", true, "true")
	convertOperator(t, "boolean", "string", true, "true")

	convertOperator(t, "primitive", "string", false, "false")
	convertOperator(t, "boolean", "string", false, "false")

	convertOperator(t, "primitive", "string", core.Binary{65, 66, 67, 226, 130, 172}, "ABC€")
	convertOperator(t, "binary", "string", core.Binary{65, 66, 67, 226, 130, 172}, "ABC€")

	convertOperator(t, "primitive", "string", nil, nil)
}
func Test_Convert__To_Boolean(t *testing.T) {
	convertOperator(t, "primitive", "boolean", "2.1", false)
	convertOperator(t, "string", "boolean", "2.1", false)

	convertOperator(t, "primitive", "boolean", 2.1, true)
	convertOperator(t, "number", "boolean", 2.1, true)

	convertOperator(t, "primitive", "boolean", 0.0, false)
	convertOperator(t, "number", "boolean", 0.0, false)

	convertOperator(t, "primitive", "boolean", "asd", false)
	convertOperator(t, "string", "boolean", "asd", false)

	convertOperator(t, "primitive", "boolean", "0", false)
	convertOperator(t, "string", "boolean", "0", false)

	convertOperator(t, "primitive", "boolean", "1", true)
	convertOperator(t, "string", "boolean", "1", true)

	convertOperator(t, "primitive", "boolean", "", false)
	convertOperator(t, "string", "boolean", "", false)

	convertOperator(t, "primitive", "boolean", core.Binary{65, 66, 67, 226, 130, 172}, false)
	convertOperator(t, "binary", "boolean", core.Binary{65, 66, 67, 226, 130, 172}, false)

	convertOperator(t, "primitive", "boolean", core.Binary{}, false)
	convertOperator(t, "binary", "boolean", core.Binary{}, false)

	convertOperator(t, "primitive", "boolean", nil, nil)
}

func Test_Convert__To_Binary(t *testing.T) {
	convertOperator(t, "primitive", "binary", "2.1", core.Binary("2.1"))
	convertOperator(t, "string", "binary", "2.1", core.Binary("2.1"))

	convertOperator(t, "primitive", "binary", 2.1, core.Binary{0xcd, 0xcc, 0xcc, 0xcc, 0xcc, 0xcc, 0x0, 0x40})
	convertOperator(t, "number", "binary", 2.1, core.Binary{0xcd, 0xcc, 0xcc, 0xcc, 0xcc, 0xcc, 0x0, 0x40})

	convertOperator(t, "primitive", "binary", true, core.Binary("true"))
	convertOperator(t, "boolean", "binary", true, core.Binary("true"))

	convertOperator(t, "primitive", "binary", false, core.Binary("false"))
	convertOperator(t, "boolean", "binary", false, core.Binary("false"))

	convertOperator(t, "primitive", "binary", core.Binary{0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40}, core.Binary{0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40})
	convertOperator(t, "binary", "binary", core.Binary{0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40}, core.Binary{0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40})

	convertOperator(t, "primitive", "binary", nil, nil)
}
