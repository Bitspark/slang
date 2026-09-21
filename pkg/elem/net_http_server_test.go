package elem

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
	"github.com/stretchr/testify/require"
)

func Test_HTTP__IsRegistered(t *testing.T) {
	a := assertions.New(t)

	ocFork := getBuiltinCfg(netHTTPServerId)
	a.NotNil(ocFork)
}

func Test_HTTP__InPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: netHTTPServerId,
		},
	)
	require.NoError(t, err)

	a.NotNil(o.Main().In())
	a.Equal(core.TYPE_NUMBER, o.Main().In().Type())
}

func Test_HTTP__OutPorts(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: netHTTPServerId,
		},
	)
	require.NoError(t, err)

	a.NotNil(o.Main().Out())
	a.Equal(core.TYPE_STRING, o.Main().Out().Type())
}

func Test_HTTP__Delegates(t *testing.T) {
	a := assertions.New(t)

	o, err := buildOperator(
		core.InstanceDef{
			Operator: netHTTPServerId,
		},
	)
	require.NoError(t, err)

	dlg := o.Delegate("handler")
	a.NotNil(dlg)

	a.Equal(core.TYPE_MAP, dlg.In().Type())
	a.Equal(core.TYPE_MAP, dlg.Out().Type())

	a.Equal(core.TYPE_BINARY, dlg.In().Map("body").Type())
	a.Equal(core.TYPE_NUMBER, dlg.In().Map("status").Type())
	a.Equal(core.TYPE_STREAM, dlg.In().Map("headers").Type())
	a.Equal(core.TYPE_MAP, dlg.In().Map("headers").Stream().Type())
	a.Equal(core.TYPE_STRING, dlg.In().Map("headers").Stream().Map("key").Type())
	a.Equal(core.TYPE_STRING, dlg.In().Map("headers").Stream().Map("value").Type())

	a.Equal(core.TYPE_STRING, dlg.Out().Map("method").Type())
	a.Equal(core.TYPE_STRING, dlg.Out().Map("path").Type())
	a.Equal(core.TYPE_STREAM, dlg.Out().Map("headers").Type())
	a.Equal(core.TYPE_MAP, dlg.Out().Map("headers").Stream().Type())
	a.Equal(core.TYPE_STRING, dlg.Out().Map("headers").Stream().Map("key").Type())
	a.Equal(core.TYPE_STREAM, dlg.Out().Map("headers").Stream().Map("values").Type())
	a.Equal(core.TYPE_STRING, dlg.Out().Map("headers").Stream().Map("values").Stream().Type())
	a.Equal(core.TYPE_MAP, dlg.Out().Map("params").Stream().Type())
	a.Equal(core.TYPE_STRING, dlg.Out().Map("params").Stream().Map("key").Type())
	a.Equal(core.TYPE_STREAM, dlg.Out().Map("params").Stream().Map("values").Type())
	a.Equal(core.TYPE_STRING, dlg.Out().Map("params").Stream().Map("values").Stream().Type())
}

func startHTTPTestOperator(t *testing.T) (*core.Operator, string) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())
	o, err := buildOperator(core.InstanceDef{Operator: netHTTPServerId})
	require.NoError(t, err)
	o.Main().Out().Bufferize()
	o.Delegate("handler").Out().Bufferize()
	o.Start()
	t.Cleanup(o.Stop)
	o.Main().In().Push(port)
	return o, fmt.Sprintf("http://127.0.0.1:%d", port)
}

func getHTTPTestResponse(url string) (*http.Response, error) {
	client := &http.Client{Timeout: 2 * time.Second}
	var err error
	for i := 0; i < 20; i++ {
		var response *http.Response
		response, err = client.Get(url)
		if err == nil {
			return response, nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil, err
}

func Test_HTTP__Request(t *testing.T) {
	a := assertions.New(t)
	o, url := startHTTPTestOperator(t)
	handler := o.Delegate("handler")
	handler.In().Push(map[string]interface{}{"status": 200, "headers": []interface{}{}, "body": core.Binary("")})
	done := make(chan error, 1)
	go func() {
		response, err := getHTTPTestResponse(url + "/test123?a=1")
		if response != nil {
			response.Body.Close()
		}
		done <- err
	}()
	a.Equal("GET", handler.Out().Map("method").Pull())
	a.Equal("/test123", handler.Out().Map("path").Pull())
	a.Equal([]interface{}{map[string]interface{}{"key": "a", "values": []interface{}{"1"}}}, handler.Out().Map("params").Pull())
	require.NoError(t, <-done)
}

func Test_HTTP__Responses(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusNotFound} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			o, url := startHTTPTestOperator(t)
			o.Delegate("handler").In().Push(map[string]interface{}{
				"status": status, "headers": []interface{}{}, "body": core.Binary("hello slang!"),
			})
			response, err := getHTTPTestResponse(url + "/test789")
			require.NoError(t, err)
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			require.Equal(t, status, response.StatusCode)
			require.Equal(t, "hello slang!", string(body))
		})
	}
}
