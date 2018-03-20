package builtin

import (
	"testing"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/stretchr/testify/require"
	"net/http"
	"bytes"
	"time"
)

func TestBuiltin_HTTP__CreatorFuncIsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg("slang.net.httpServer")
	a.NotNil(ocFork)
}

func TestBuiltin_HTTP__InPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := MakeOperator(
		core.InstanceDef{
			Operator: "slang.net.httpServer",
		},
	)
	require.NoError(t, err)

	a.NotNil(o.DefaultService().In())
	a.Equal(core.TYPE_NUMBER, o.DefaultService().In().Type())
}

func TestBuiltin_HTTP__OutPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := MakeOperator(
		core.InstanceDef{
			Operator: "slang.net.httpServer",
		},
	)
	require.NoError(t, err)

	a.NotNil(o.DefaultService().Out())
	a.Equal(core.TYPE_STRING, o.DefaultService().Out().Type())
}

func TestBuiltin_HTTP__Delegates(t *testing.T) {
	a := assertions.New(t)

	o, err := MakeOperator(
		core.InstanceDef{
			Operator: "slang.net.httpServer",
		},
	)
	require.NoError(t, err)

	dlg := o.Delegate("handler")
	a.NotNil(dlg)

	a.Equal(core.TYPE_STREAM, dlg.In().Type())
	a.Equal(core.TYPE_STREAM, dlg.Out().Type())

	a.Equal(core.TYPE_MAP, dlg.In().Stream().Type())
	a.Equal(core.TYPE_MAP, dlg.Out().Stream().Type())

	a.Equal(core.TYPE_BINARY, dlg.In().Stream().Map("body").Type())
	a.Equal(core.TYPE_NUMBER, dlg.In().Stream().Map("status").Type())
	a.Equal(core.TYPE_STREAM, dlg.In().Stream().Map("headers").Type())
	a.Equal(core.TYPE_MAP, dlg.In().Stream().Map("headers").Stream().Type())
	a.Equal(core.TYPE_STRING, dlg.In().Stream().Map("headers").Stream().Map("key").Type())
	a.Equal(core.TYPE_STRING, dlg.In().Stream().Map("headers").Stream().Map("value").Type())

	a.Equal(core.TYPE_STRING, dlg.Out().Stream().Map("method").Type())
	a.Equal(core.TYPE_STRING, dlg.Out().Stream().Map("path").Type())
	a.Equal(core.TYPE_STRING, dlg.Out().Stream().Map("protocol").Type())
	a.Equal(core.TYPE_STREAM, dlg.Out().Stream().Map("headers").Type())
	a.Equal(core.TYPE_MAP, dlg.Out().Stream().Map("headers").Stream().Type())
	a.Equal(core.TYPE_STRING, dlg.Out().Stream().Map("headers").Stream().Map("key").Type())
	a.Equal(core.TYPE_STRING, dlg.Out().Stream().Map("headers").Stream().Map("value").Type())
}

func TestBuiltin_HTTP__Request(t *testing.T) {
	a := assertions.New(t)

	o, err := MakeOperator(
		core.InstanceDef{
			Operator: "slang.net.httpServer",
		},
	)
	require.NoError(t, err)

	o.DefaultService().Out().Bufferize()
	handler := o.Delegate("handler")
	handler.Out().Bufferize()

	o.Start()
	o.DefaultService().In().Push(9438)
	a.True(handler.Out().PullBOS())
	handler.In().PushBOS()

	done := false

	go func() {
		for i := 0; i < 5; i++ {
			http.Get("http://127.0.0.1:9438/test123")
			if done {
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
	}()

	a.Equal("GET", handler.Out().Stream().Map("method").Pull())
	a.Equal("/test123", handler.Out().Stream().Map("path").Pull())
	done = true
}

func TestBuiltin_HTTP__Response200(t *testing.T) {
	a := assertions.New(t)

	o, err := MakeOperator(
		core.InstanceDef{
			Operator: "slang.net.httpServer",
		},
	)
	require.NoError(t, err)

	o.DefaultService().Out().Bufferize()
	handler := o.Delegate("handler")
	handler.Out().Bufferize()

	o.Start()
	o.DefaultService().In().Push(9439)
	a.True(handler.Out().PullBOS())
	handler.In().PushBOS()
	handler.In().Stream().Push(map[string]interface{}{"status": 200, "headers": []interface{}{}, "body": []byte("hallo slang!")})

	for i := 0; i < 5; i++ {
		resp, _ := http.Get("http://127.0.0.1:9439/test789")
		if resp == nil || resp.StatusCode != 200 {
			time.Sleep(20 * time.Millisecond)
			continue
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		a.Equal([]byte("hallo slang!"), buf.Bytes())
		a.Equal("200 OK", resp.Status)
		return
	}
	a.Fail("no response")
}

func TestBuiltin_HTTP__Response404(t *testing.T) {
	a := assertions.New(t)

	o, err := MakeOperator(
		core.InstanceDef{
			Operator: "slang.net.httpServer",
		},
	)
	require.NoError(t, err)

	o.DefaultService().Out().Bufferize()
	handler := o.Delegate("handler")
	handler.Out().Bufferize()

	o.Start()
	o.DefaultService().In().Push(9440)
	a.True(handler.Out().PullBOS())
	handler.In().PushBOS()
	handler.In().Stream().Push(map[string]interface{}{"status": 404, "headers": []interface{}{}, "body": []byte("bye slang!")})

	for i := 0; i < 5; i++ {
		resp, _ := http.Get("http://127.0.0.1:9440/test789")
		if resp == nil || resp.StatusCode != 404 {
			time.Sleep(20 * time.Millisecond)
			continue
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		a.Equal([]byte("bye slang!"), buf.Bytes())
		a.Equal("404 Not Found", resp.Status)
		return
	}
	a.Fail("no response")
}
