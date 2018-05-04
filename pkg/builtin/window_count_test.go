package builtin

import (
	"testing"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/stretchr/testify/require"
	"github.com/Bitspark/slang/pkg/core"
)

func TestOperatorWindow__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocWnd := getBuiltinCfg("slang.window.count")
	a.NotNil(ocWnd)
}

func makeTestMonoWindow(t *testing.T, size, slide, start, end int) *core.Operator {
	r := require.New(t)
	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.window.count",
			Generics: map[string]*core.TypeDef{
				"itemType": {
					Type: "string",
				},
			},
			Properties: map[string]interface{}{
				"size":  size,  // maximum and normal size of ordinary windows
				"slide": slide, // distance between two first elements of two consecutive windows
				"start": start, // minimum size of the first window
				"end":   end,   // minimum size of the last window
			},
		},
	)
	r.NoError(err)
	r.NotNil(o)
	o.Main().Out().Bufferize()
	o.Start()
	return o
}

func TestOperatorWindowCount_3_3_3_3(t *testing.T) {
	a := assertions.New(t)
	o := makeTestMonoWindow(t, 3, 3, 3, 3)

	o.Main().In().Push([]interface{}{"a", "b", "c", "d", "e", "f", "g", "h", "i"})

	o.Main().Out().PullBOS()
	a.PortPushes([]interface{}{"a", "b", "c"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"d", "e", "f"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"g", "h", "i"}, o.Main().Out().Stream())
	o.Main().Out().PullEOS()
}

func TestOperatorWindowCount_3_3_1_3(t *testing.T) {
	a := assertions.New(t)
	o := makeTestMonoWindow(t, 3, 3, 1, 3)

	o.Main().In().Push([]interface{}{"a", "b", "c", "d", "e", "f", "g", "h", "i"})

	o.Main().Out().PullBOS()
	a.PortPushes([]interface{}{"a"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"b", "c", "d"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"e", "f", "g"}, o.Main().Out().Stream())
	o.Main().Out().PullEOS()
}

func TestOperatorWindowCount_3_3_1_2(t *testing.T) {
	a := assertions.New(t)
	o := makeTestMonoWindow(t, 3, 3, 1, 2)

	o.Main().In().Push([]interface{}{"a", "b", "c", "d", "e", "f", "g", "h", "i"})

	o.Main().Out().PullBOS()
	a.PortPushes([]interface{}{"a"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"b", "c", "d"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"e", "f", "g"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"h", "i"}, o.Main().Out().Stream())
	o.Main().Out().PullEOS()
}

func TestOperatorWindowCount_3_3_2_1(t *testing.T) {
	a := assertions.New(t)
	o := makeTestMonoWindow(t, 3, 3, 2, 1)

	o.Main().In().Push([]interface{}{"a", "b", "c", "d", "e", "f", "g", "h", "i"})

	o.Main().Out().PullBOS()
	a.PortPushes([]interface{}{"a", "b"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"c", "d", "e"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"f", "g", "h"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"i"}, o.Main().Out().Stream())
	o.Main().Out().PullEOS()
}

func TestOperatorWindowCount_3_2_2_2(t *testing.T) {
	a := assertions.New(t)
	o := makeTestMonoWindow(t, 3, 2, 2, 2)

	o.Main().In().Push([]interface{}{"a", "b", "c", "d", "e", "f", "g", "h", "i"})

	o.Main().Out().PullBOS()
	a.PortPushes([]interface{}{"a", "b"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"b", "c", "d"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"d", "e", "f"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"f", "g", "h"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"h", "i"}, o.Main().Out().Stream())
	o.Main().Out().PullEOS()
}

func TestOperatorWindowCount_3_2_2_3(t *testing.T) {
	a := assertions.New(t)
	o := makeTestMonoWindow(t, 3, 2, 2, 3)

	o.Main().In().Push([]interface{}{"a", "b", "c", "d", "e", "f", "g", "h", "i"})

	o.Main().Out().PullBOS()
	a.PortPushes([]interface{}{"a", "b"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"b", "c", "d"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"d", "e", "f"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"f", "g", "h"}, o.Main().Out().Stream())
	o.Main().Out().PullEOS()
}

func TestOperatorWindowCount_3_1_2_2(t *testing.T) {
	a := assertions.New(t)
	o := makeTestMonoWindow(t, 3, 1, 2, 2)

	o.Main().In().Push([]interface{}{"a", "b", "c", "d", "e", "f", "g", "h", "i"})

	o.Main().Out().PullBOS()
	a.PortPushes([]interface{}{"a", "b"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"a", "b", "c"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"b", "c", "d"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"c", "d", "e"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"d", "e", "f"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"e", "f", "g"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"f", "g", "h"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"g", "h", "i"}, o.Main().Out().Stream())
	a.PortPushes([]interface{}{"h", "i"}, o.Main().Out().Stream())
	o.Main().Out().PullEOS()
}
