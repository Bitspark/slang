package daemon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestServer_operator(t *testing.T) {
	server := getTestServer()
	request, _ := http.NewRequest("GET", "/operator/", nil)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code)
}
