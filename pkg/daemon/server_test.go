package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/env"
	"github.com/Bitspark/slang/pkg/storage"
	"github.com/Bitspark/slang/tests"
	"github.com/stretchr/testify/assert"
)

func getTestServer() *Server {
	backend := tests.NewTestLoader("./")
	env := env.New("localhost", 8000)
	storage := storage.NewStorage().AddBackend(backend)
	ctx := SetStorage(context.Background(), storage)
	return NewServer(&ctx, env)
}
func TestServer_operator_starting(t *testing.T) {
	server := getTestServer()
	data := runInstructionJSON{Id: "37ccdc28-67b0-4bb1-8591-4e0e813e3ec1",
		Stream: false,
		Props:  core.Properties{}, Gens: core.Generics{}}
	body, _ := json.Marshal(&data)
	request, _ := http.NewRequest("POST", "/run/", bytes.NewReader(body))
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	a := response.Body.String()
	fmt.Println(a)
	assert.Equal(t, 200, response.Code)
}

func TestServer_operator(t *testing.T) {
	server := getTestServer()
	request, _ := http.NewRequest("GET", "/operator/", nil)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code)
}
