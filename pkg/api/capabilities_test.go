package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// The HTTP client operator's blueprint identifier.
var httpClientOperator = uuid.MustParse("f7f5907d-758b-4892-8a3e-ae86b877b869")

type recordingTransport struct{ requests int64 }

func (r *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	atomic.AddInt64(&r.requests, 1)
	return http.DefaultTransport.RoundTrip(req)
}

// httpBundle wraps one HTTP client operator in a program with the same interface.
func httpBundle(t *testing.T) *core.SlangBundle {
	t.Helper()
	elem.Init()
	operator, err := elem.GetBlueprint(httpClientOperator)
	require.NoError(t, err)
	service := operator.ServiceDefs[core.MAIN_SERVICE]
	id := uuid.New()
	program := core.Blueprint{
		Id: id,
		ServiceDefs: map[string]*core.ServiceDef{core.MAIN_SERVICE: {
			In: service.In.Copy(), Out: service.Out.Copy(),
		}},
		InstanceDefs: core.InstanceDefList{{Name: "http", Operator: httpClientOperator}},
		Connections:  map[string][]string{"(": {"(http"}, "http)": {")"}},
	}
	return &core.SlangBundle{Main: id, Blueprints: map[uuid.UUID]core.Blueprint{id: program}}
}

// Compile flattens a program and rebuilds every elementary operator; the rebuilt
// operators must keep the capabilities the program was built with.
func TestBuildOperatorWithKeepsCapabilitiesThroughCompile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, "ok")
	}))
	defer server.Close()
	transport := &recordingTransport{}

	op, err := BuildOperatorWith(httpBundle(t), elem.Capabilities{HTTP: &http.Client{Transport: transport}})
	require.NoError(t, err)
	op.Main().Out().Bufferize()
	go op.Start()
	defer op.Stop()

	op.Main().In().Push(map[string]interface{}{
		"method": "GET", "url": server.URL, "headers": []interface{}{}, "body": core.Binary{},
	})
	response := op.Main().Out().Pull().(map[string]interface{})

	require.Equal(t, 200.0, response["status"])
	require.Equal(t, int64(1), atomic.LoadInt64(&transport.requests), "the request must use the injected transport")
}

func TestBuildOperatorWithRefusesAnOperatorWhoseCapabilityIsMissing(t *testing.T) {
	_, err := BuildOperatorWith(httpBundle(t), elem.Capabilities{})
	require.ErrorContains(t, err, "does not provide HTTP")
}
