package daemon

import (
	"net/http"
	"testing"
	"net/http/httptest"
	"github.com/Bitspark/slang/tests/assertions"
	"encoding/json"
	"bytes"
)

func Test_ServiceOperatorDef_Endpoint_GET__SimpleOperator(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	inData, _ := json.Marshal(&inJSON{
		WorkingDir: "tests/test_data/daemon/services/operator",
	})

	r, _ := http.NewRequest("GET", "/", bytes.NewReader(inData))
	w := httptest.NewRecorder()

	OperatorDefService.Routes["/"].Handle(w, r)

	var outData outJSON
	json.Unmarshal(w.Body.Bytes(), &outData)

	a.Equal("success", outData.Status)
	a.Empty(outData.Error)
	a.NotEmpty(outData.Objects)
}
