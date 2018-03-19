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
	o, err := MakeOperator(
		core.InstanceDef{
			Operator: "slang.window.count",
			Generics: map[string]*core.PortDef{
				"inStreams": {
					Type: "stream",
					Stream: &core.PortDef{
						Type: "string",
					},
				},
				"outStreams": {
					Type: "string",
				},
			},
			Properties: map[string]interface{}{
				"size":  float64(size),  // maximum and normal size of ordinary windows
				"slide": float64(slide), // distance between two first elements of two consecutive windows
				"start": float64(start), // minimum size of the first window
				"end":   float64(end),   // minimum size of the last window
			},
		},
	)
	r.NotNil(o)
	r.NoError(err)
	o.Out().Bufferize()
	o.Start()
	return o
}

/*func TestOperatorWindow__Stereo_Tumbling(t *testing.T) {
	a := assertions.New(t)
	o, err := MakeOperator(
		core.InstanceDef{
			Operator: "slang.window.count",
			Generics: map[string]*core.PortDef{
				"inStreams": {
					Type: "map",
					Map: map[string]*core.PortDef{
						"a": {
							Type: "stream",
							Stream: &core.PortDef{
								Type: "string",
							},
						},
						"b": {
							Type: "stream",
							Stream: &core.PortDef{
								Type: "number",
							},
						},
					},
				},
				"outStreams": {
					Type: "map",
					Map: map[string]*core.PortDef{
						"a": {
							Type: "string",
						},
						"b": {
							Type: "number",
						},
					},
				},
			},
			Properties: map[string]interface{}{
				"size":  float64(3), // maximum and normal size of ordinary windows
				"slide": float64(3), // distance between two first elements of two consecutive windows
				"start": float64(3), // minimum size of the first window
				"end":   float64(3), // minimum size of the last window
			},
		},
	)
	require.NoError(t, err)
	o.Out().Bufferize()

	o.Start()

	o.In().Push([]interface{}{
		map[string]interface{}{"a": "a", "b": 1},
		map[string]interface{}{"a": "b", "b": 2},
		map[string]interface{}{"a": "c", "b": 3},
		map[string]interface{}{"a": "d", "b": 4},
		map[string]interface{}{"a": "e", "b": 5},
		map[string]interface{}{"a": "f", "b": 6},
	})

	o.Out().PullBOS()
	a.PortPushes([]interface{}{
		map[string]interface{}{"a": "a", "b": 1},
		map[string]interface{}{"a": "b", "b": 2},
		map[string]interface{}{"a": "c", "b": 3},
	}, o.Out().Stream())
	a.PortPushes([]interface{}{
		map[string]interface{}{"a": "d", "b": 4},
		map[string]interface{}{"a": "e", "b": 5},
		map[string]interface{}{"a": "f", "b": 6},
	}, o.Out().Stream())
	o.Out().PullEOS()
}*/

func TestOperatorWindow__Mono_3_3_3_3(t *testing.T) {
	a := assertions.New(t)
	o := makeTestMonoWindow(t, 3, 3, 3, 3)

	o.In().Push([]interface{}{"a", "b", "c", "d", "e", "f", "g", "h", "i"})

	o.Out().PullBOS()
	a.PortPushes([]interface{}{"a", "b", "c"}, o.Out().Stream())
	a.PortPushes([]interface{}{"d", "e", "f"}, o.Out().Stream())
	a.PortPushes([]interface{}{"g", "h", "i"}, o.Out().Stream())
	o.Out().PullEOS()
}

func TestOperatorWindow__Mono_3_3_1_3(t *testing.T) {
	a := assertions.New(t)
	o := makeTestMonoWindow(t, 3, 3, 1, 3)

	o.In().Push([]interface{}{"a", "b", "c", "d", "e", "f", "g", "h", "i"})

	o.Out().PullBOS()
	a.PortPushes([]interface{}{"a"}, o.Out().Stream())
	a.PortPushes([]interface{}{"b", "c", "d"}, o.Out().Stream())
	a.PortPushes([]interface{}{"e", "f", "g"}, o.Out().Stream())
	o.Out().PullEOS()
}

func TestOperatorWindow__Mono_3_3_1_2(t *testing.T) {
	a := assertions.New(t)
	o := makeTestMonoWindow(t, 3, 3, 1, 2)

	o.In().Push([]interface{}{"a", "b", "c", "d", "e", "f", "g", "h", "i"})

	o.Out().PullBOS()
	a.PortPushes([]interface{}{"a"}, o.Out().Stream())
	a.PortPushes([]interface{}{"b", "c", "d"}, o.Out().Stream())
	a.PortPushes([]interface{}{"e", "f", "g"}, o.Out().Stream())
	a.PortPushes([]interface{}{"h", "i"}, o.Out().Stream())
	o.Out().PullEOS()
}

func TestOperatorWindow__Mono_3_3_2_1(t *testing.T) {
	a := assertions.New(t)
	o := makeTestMonoWindow(t, 3, 3, 2, 1)

	o.In().Push([]interface{}{"a", "b", "c", "d", "e", "f", "g", "h", "i"})

	o.Out().PullBOS()
	a.PortPushes([]interface{}{"a", "b"}, o.Out().Stream())
	a.PortPushes([]interface{}{"c", "d", "e"}, o.Out().Stream())
	a.PortPushes([]interface{}{"f", "g", "h"}, o.Out().Stream())
	a.PortPushes([]interface{}{"i"}, o.Out().Stream())
	o.Out().PullEOS()
}

func TestOperatorWindow__Mono_3_2_2_2(t *testing.T) {
	a := assertions.New(t)
	o := makeTestMonoWindow(t, 3, 2, 2, 2)

	o.In().Push([]interface{}{"a", "b", "c", "d", "e", "f", "g", "h", "i"})

	o.Out().PullBOS()
	a.PortPushes([]interface{}{"a", "b"}, o.Out().Stream())
	a.PortPushes([]interface{}{"b", "c", "d"}, o.Out().Stream())
	a.PortPushes([]interface{}{"d", "e", "f"}, o.Out().Stream())
	a.PortPushes([]interface{}{"f", "g", "h"}, o.Out().Stream())
	a.PortPushes([]interface{}{"h", "i"}, o.Out().Stream())
	o.Out().PullEOS()
}

func TestOperatorWindow__Mono_3_2_2_3(t *testing.T) {
	a := assertions.New(t)
	o := makeTestMonoWindow(t, 3, 2, 2, 3)

	o.In().Push([]interface{}{"a", "b", "c", "d", "e", "f", "g", "h", "i"})

	o.Out().PullBOS()
	a.PortPushes([]interface{}{"a", "b"}, o.Out().Stream())
	a.PortPushes([]interface{}{"b", "c", "d"}, o.Out().Stream())
	a.PortPushes([]interface{}{"d", "e", "f"}, o.Out().Stream())
	a.PortPushes([]interface{}{"f", "g", "h"}, o.Out().Stream())
	o.Out().PullEOS()
}

func TestOperatorWindow__Mono_3_1_2_2(t *testing.T) {
	a := assertions.New(t)
	o := makeTestMonoWindow(t, 3, 1, 2, 2)

	o.In().Push([]interface{}{"a", "b", "c", "d", "e", "f", "g", "h", "i"})

	o.Out().PullBOS()
	a.PortPushes([]interface{}{"a", "b"}, o.Out().Stream())
	a.PortPushes([]interface{}{"a", "b", "c"}, o.Out().Stream())
	a.PortPushes([]interface{}{"b", "c", "d"}, o.Out().Stream())
	a.PortPushes([]interface{}{"c", "d", "e"}, o.Out().Stream())
	a.PortPushes([]interface{}{"d", "e", "f"}, o.Out().Stream())
	a.PortPushes([]interface{}{"e", "f", "g"}, o.Out().Stream())
	a.PortPushes([]interface{}{"f", "g", "h"}, o.Out().Stream())
	a.PortPushes([]interface{}{"g", "h", "i"}, o.Out().Stream())
	a.PortPushes([]interface{}{"h", "i"}, o.Out().Stream())
	o.Out().PullEOS()
}
