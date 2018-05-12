package daemon

import (
	"net/http"
	"testing"
	"net/http/httptest"
	"github.com/Bitspark/slang/tests/assertions"
	"encoding/json"
)

func Test_ServiceOperatorDef_Endpoint_GET__SimpleOperator(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	r, _ := http.NewRequest("GET", "/", nil)
	q := r.URL.Query()
	q.Add("cwd", "../../tests/test_data/daemon/services/operator")
	r.URL.RawQuery = q.Encode()
	w := httptest.NewRecorder()

	OperatorDefService.Routes["/"].Handle(w, r)

	var outData outJSON
	json.Unmarshal(w.Body.Bytes(), &outData)

	a.Empty(outData.Error)
	a.True(0 < len(outData.Objects))
}
