package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func getTestServer() *httptest.Server {
	backend := NewTestLoader("./")
	env := env.New("localhost", 8000)
	storage := storage.NewStorage().AddBackend(backend)
	ctx := daemon.SetStorage(context.Background(), storage)
	backend.Reload()
	s := daemon.NewServer(&ctx, env)
	return httptest.NewServer(s.Handler())
}

func startOperator(t *testing.T, s *httptest.Server, ri daemon.RunInstruction) daemon.RunState {
	var out daemon.RunState

	body, _ := json.Marshal(&ri)
	request, _ := http.NewRequest("POST", s.URL+"/run/", bytes.NewReader(body))
	response, err := s.Client().Do(request)
	if err != nil {
		fmt.Println(err)
	}
	body, _ = ioutil.ReadAll(response.Body)
	json.Unmarshal(body, &out)

	assert.Equal(t, 200, response.StatusCode)
	assert.Equal(t, "success", out.Status)
	return out
}

func TestServer_operator_starting(t *testing.T) {
	server := getTestServer()
	id, _ := uuid.Parse("8b62495a-e482-4a3e-8020-0ab8a350ad2d")
	data := daemon.RunInstruction{Id: id,
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
	c := server.Client()
	response, _ := c.Do(request)
	assert.Equal(t, "\"slang\"", response.Body)
}

func TestServer_operator(t *testing.T) {
	server := getTestServer()
	id := "8b62495a-e482-4a3e-8020-0ab8a350ad2d"
	uuid, _ := uuid.Parse("8b62495a-e482-4a3e-8020-0ab8a350ad2d")
	data := daemon.RunInstruction{Id: uuid,
		Stream: false,
		Props:  core.Properties{"value": "slang"},
		Gens: core.Generics{
			"valueType": {
				Type: "string",
			},
		},
	}
	startOperator(t, server, data)
	request, _ := http.NewRequest("GET", server.URL+"/instances/", nil)
	response, _ := server.Client().Do(request)
	assert.Equal(t, 200, response.StatusCode)
	body, _ := ioutil.ReadAll(response.Body)
	assert.Contains(t, string(body), id)
}
