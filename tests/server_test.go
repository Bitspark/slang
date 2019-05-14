package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/daemon"
	"github.com/Bitspark/slang/pkg/env"
	"github.com/Bitspark/slang/pkg/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func getTestServer() *daemon.Server {
	backend := NewTestLoader("./")
	env := env.New("localhost", 8000)
	storage := storage.NewStorage().AddBackend(backend)
	ctx := daemon.SetStorage(context.Background(), storage)
	backend.Reload()
	return daemon.NewServer(&ctx, env)
}

func startOperator(t *testing.T, s *daemon.Server, ri daemon.RunInstructionJSON) daemon.InstanceStateJSON {
	var out daemon.InstanceStateJSON

	body, _ := json.Marshal(&ri)
	request, _ := http.NewRequest("POST", "/run/", bytes.NewReader(body))
	response := httptest.NewRecorder()
	s.Handler().ServeHTTP(response, request)

	decoder := json.NewDecoder(response.Body)
	decoder.Decode(&out)

	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "success", out.Status)
	return out
}
func TestServer_operator_starting(t *testing.T) {
	server := getTestServer()
	id, _ := uuid.Parse("8b62495a-e482-4a3e-8020-0ab8a350ad2d")
	data := daemon.RunInstructionJSON{Id: id,
		Stream: false,
		Props:  core.Properties{"value": "slang"},
		Gens: core.Generics{
			"valueType": {
				Type: "string",
			},
		},
	}
	instance := startOperator(t, server, data)
	request, _ := http.NewRequest("POST", strings.Join([]string{"/instance/", instance.Handle}, ""), bytes.NewReader([]byte{}))
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	assert.Equal(t, "\"slang\"", response.Body.String())
}

func TestServer_operator(t *testing.T) {
	server := getTestServer()
	id := "8b62495a-e482-4a3e-8020-0ab8a350ad2d"
	uuid, _ := uuid.Parse("8b62495a-e482-4a3e-8020-0ab8a350ad2d")
	data := daemon.RunInstructionJSON{Id: uuid,
		Stream: false,
		Props:  core.Properties{"value": "slang"},
		Gens: core.Generics{
			"valueType": {
				Type: "string",
			},
		},
	}
	startOperator(t, server, data)
	request, _ := http.NewRequest("GET", "/instances/", nil)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code)
	assert.Contains(t, response.Body.String(), id)
}
