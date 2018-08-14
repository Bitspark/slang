package daemon

import (
	"net/http"
	"testing"
	"net/http/httptest"
	"github.com/Bitspark/slang/tests/assertions"
	"encoding/json"
	"github.com/Bitspark/slang/pkg/api"
)

var TEST_ENV = api.NewTestEnviron("../../tests/test_data/daemon/services/operator")

func Test_ServiceOperatorDef_Endpoint_GET__SimpleOperator(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	r, _ := http.NewRequest("GET", "/", nil)
	q := r.URL.Query()
	r.URL.RawQuery = q.Encode()
	w := httptest.NewRecorder()

	DefinitionService.Routes["/"].Handle(TEST_ENV, w, r)

	var outData struct {
		Objects []interface{} `json:"objects"`
		Status  string        `json:"status"`
		Error   *Error        `json:"error,omitempty"`
	}
	json.Unmarshal(w.Body.Bytes(), &outData)

	a.Empty(outData.Error)
	a.True(0 < len(outData.Objects))
}
