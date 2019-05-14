package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/daemon"
	"github.com/Bitspark/slang/pkg/env"
	"github.com/Bitspark/slang/pkg/storage"
	"github.com/Bitspark/slang/tests"
	"github.com/stretchr/testify/assert"
)

func getTestServer() *daemon.Server {
	backend := tests.NewTestLoader("./")
	env := env.New("localhost", 8000)
	storage := storage.NewStorage().AddBackend(backend)
	ctx := daemon.SetStorage(context.Background(), storage)
	backend.Reload()
	fmt.Println(backend.List())
	return daemon.NewServer(&ctx, env)
}
func TestServer_operator_starting(t *testing.T) {
	server := getTestServer()
	data := daemon.RunInstructionJSON{Id: "8b62495a-e482-4a3e-8020-0ab8a350ad2d",
		Stream: false,
		Props:  core.Properties{"value": "slang"},
		Gens: core.Generics{
			"valueType": {
				Type: "string",
			},
		},
	}
	body, _ := json.Marshal(&data)
	request, _ := http.NewRequest("POST", "/run/", bytes.NewReader(body))
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	decoder := json.NewDecoder(response.Body)
	var out daemon.InstanceStateJSON
	decoder.Decode(&out)
	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "success", out.Status)

	request, _ = http.NewRequest("POST", strings.Join([]string{"/instance/", out.Handle}, ""), bytes.NewReader([]byte{}))
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	assert.Equal(t, "\"slang\"", response.Body.String())
}

func TestServer_operator(t *testing.T) {
	server := getTestServer()
	request, _ := http.NewRequest("GET", "/operator/", nil)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code)
}
