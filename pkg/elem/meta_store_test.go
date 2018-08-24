package elem

import (
	"testing"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/stretchr/testify/require"
	"time"
)

func Test_MetaStore__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocMetaStore := getBuiltinCfg("slang.meta.Store")
	a.NotNil(ocMetaStore)
}

func Test_MetaStore__String(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.meta.Store",
			Generics: map[string]*core.TypeDef{
				"examineType": {
					Type: "string",
				},
			},
		},
	)
	require.NoError(t, err)

	querySrv := o.Service("query")

	querySrv.Out().Bufferize()

	querySrv.In().Push(nil)
	a.Equal([]interface{}{}, querySrv.Out().Pull())

	o.Main().In().Push("test1")
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{"test1"}, querySrv.Out().Pull())

	o.Main().In().Push("test2")
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{"test1", "test2"}, querySrv.Out().Pull())

	o.Main().In().Push("test3")
	o.Main().In().Push("test4")
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{"test1", "test2", "test3", "test4"}, querySrv.Out().Pull())
}

func Test_MetaStore__Stream(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.meta.Store",
			Generics: map[string]*core.TypeDef{
				"examineType": {
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "number",
					},
				},
			},
		},
	)
	require.NoError(t, err)

	querySrv := o.Service("query")

	querySrv.Out().Bufferize()

	querySrv.In().Push(nil)
	a.Equal([]interface{}{}, querySrv.Out().Pull())

	o.Main().In().PushBOS()
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{[]interface{}{core.UnfinishedMarker}}, querySrv.Out().Pull())

	o.Main().In().Stream().Push(1.0)
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{[]interface{}{1.0, core.UnfinishedMarker}}, querySrv.Out().Pull())

	o.Main().In().Stream().Push(2.0)
	o.Main().In().PushEOS()
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{[]interface{}{1.0, 2.0}}, querySrv.Out().Pull())

	o.Main().In().PushBOS()
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{[]interface{}{1.0, 2.0}, []interface{}{core.UnfinishedMarker}}, querySrv.Out().Pull())

	o.Main().In().PushEOS()
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{[]interface{}{1.0, 2.0}, []interface{}{}}, querySrv.Out().Pull())
}

func Test_MetaStore__Map(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.meta.Store",
			Generics: map[string]*core.TypeDef{
				"examineType": {
					Type: "map",
					Map: map[string]*core.TypeDef{
						"a": {
							Type: "string",
						},
						"b": {
							Type: "boolean",
						},
					},
				},
			},
		},
	)
	require.NoError(t, err)

	querySrv := o.Service("query")

	querySrv.Out().Bufferize()

	querySrv.In().Push(nil)
	a.Equal([]interface{}{}, querySrv.Out().Pull())

	o.Main().In().Map("a").Push("test1")
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{
		map[string]interface{}{
			"a": "test1",
			"b": core.PlaceholderMarker,
		},
	}, querySrv.Out().Pull())

	o.Main().In().Map("b").Push(true)
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{
		map[string]interface{}{
			"a": "test1",
			"b": true,
		},
	}, querySrv.Out().Pull())

	o.Main().In().Map("a").Push("test2")
	o.Main().In().Map("b").Push(false)
	o.Main().In().PushEOS()
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{
		map[string]interface{}{
			"a": "test1",
			"b": true,
		},
		map[string]interface{}{
			"a": "test2",
			"b": false,
		},
	}, querySrv.Out().Pull())
}

func Test_MetaStore__StreamMap(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.meta.Store",
			Generics: map[string]*core.TypeDef{
				"examineType": {
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "map",
						Map: map[string]*core.TypeDef{
							"a": {
								Type: "string",
							},
							"b": {
								Type: "boolean",
							},
							"c": {
								Type: "stream",
								Stream: &core.TypeDef{
									Type: "trigger",
								},
							},
						},
					},
				},
			},
		},
	)
	require.NoError(t, err)

	querySrv := o.Service("query")

	querySrv.Out().Bufferize()

	querySrv.In().Push(nil)
	a.Equal([]interface{}{}, querySrv.Out().Pull())

	o.Main().In().Stream().Map("a").Push(core.BOS{o.Main().In()})
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{[]interface{}{core.UnfinishedMarker}}, querySrv.Out().Pull())

	o.Main().In().Stream().Map("a").Push("test1")
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{[]interface{}{
		map[string]interface{}{
			"a": "test1",
			"b": core.PlaceholderMarker,
			"c": core.PlaceholderMarker,
		},
	}}, querySrv.Out().Pull())

	o.Main().In().Stream().Map("b").Push(core.BOS{o.Main().In()})
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{[]interface{}{
		map[string]interface{}{
			"a": "test1",
			"b": core.PlaceholderMarker,
			"c": core.PlaceholderMarker,
		},
	}}, querySrv.Out().Pull())

	o.Main().In().Stream().Map("c").Push(core.BOS{o.Main().In()})
	o.Main().In().Stream().Map("c").Push(core.BOS{o.Main().In().Stream().Map("c")})
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{[]interface{}{
		map[string]interface{}{
			"a": "test1",
			"b": core.PlaceholderMarker,
			"c": []interface{}{core.UnfinishedMarker},
		},
	}}, querySrv.Out().Pull())
}

func Test_MetaStore__MapStream(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: "slang.meta.Store",
			Generics: map[string]*core.TypeDef{
				"examineType": {
					Type: "map",
					Map: map[string]*core.TypeDef{
						"a": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type: "string",
							},
						},
						"b": {
							Type: "stream",
							Stream: &core.TypeDef{
								Type: "boolean",
							},
						},
						"c": {
							Type: "map",
							Map: map[string]*core.TypeDef{
								"a": {
									Type: "trigger",
								},
								"d": {
									Type: "number",
								},
							},
						},
					},
				},
			},
		},
	)
	require.NoError(t, err)

	querySrv := o.Service("query")

	querySrv.Out().Bufferize()

	querySrv.In().Push(nil)
	a.Equal([]interface{}{}, querySrv.Out().Pull())

	o.Main().In().Map("b").PushBOS()
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{
		map[string]interface{}{
			"a": core.PlaceholderMarker,
			"b": []interface{}{core.UnfinishedMarker},
			"c": map[string]interface{}{
				"a": core.PlaceholderMarker,
				"b": core.PlaceholderMarker,
			},
		},
	}, querySrv.Out().Pull())

	o.Main().In().Map("b").Stream().Push(true)
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{
		map[string]interface{}{
			"a": core.PlaceholderMarker,
			"b": []interface{}{true, core.UnfinishedMarker},
			"c": map[string]interface{}{
				"a": core.PlaceholderMarker,
				"b": core.PlaceholderMarker,
			},
		},
	}, querySrv.Out().Pull())

	o.Main().In().Map("c").Map("a").Push(nil)
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{
		map[string]interface{}{
			"a": core.PlaceholderMarker,
			"b": []interface{}{true, core.UnfinishedMarker},
			"c": map[string]interface{}{
				"a": nil,
				"b": core.PlaceholderMarker,
			},
		},
	}, querySrv.Out().Pull())

	o.Main().In().Map("c").Map("a").Push(nil)
	o.Main().In().Map("b").PushEOS()
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{
		map[string]interface{}{
			"a": core.PlaceholderMarker,
			"b": []interface{}{true},
			"c": map[string]interface{}{
				"a": nil,
				"b": core.PlaceholderMarker,
			},
		},
		map[string]interface{}{
			"a": core.PlaceholderMarker,
			"b": core.PlaceholderMarker,
			"c": map[string]interface{}{
				"a": nil,
				"b": core.PlaceholderMarker,
			},
		},
	}, querySrv.Out().Pull())

	o.Main().In().Map("b").PushBOS()
	time.Sleep(20 * time.Millisecond)
	querySrv.In().Push(nil)
	a.Equal([]interface{}{
		map[string]interface{}{
			"a": core.PlaceholderMarker,
			"b": []interface{}{true},
			"c": map[string]interface{}{
				"a": nil,
				"b": core.PlaceholderMarker,
			},
		},
		map[string]interface{}{
			"a": core.PlaceholderMarker,
			"b": []interface{}{core.UnfinishedMarker},
			"c": map[string]interface{}{
				"a": nil,
				"b": core.PlaceholderMarker,
			},
		},
	}, querySrv.Out().Pull())
}
