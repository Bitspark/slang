package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/daemon"
	"github.com/Bitspark/slang/pkg/env"
	"github.com/Bitspark/slang/pkg/storage"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func getTestServer() *httptest.Server {
	env := env.New("localhost", 8000)
	st := storage.NewStorage().
		AddBackend(storage.NewReadOnlyFileSystem("../fixtures"))

	ctx := daemon.SetStorage(context.Background(), st)
	s := daemon.NewServer(&ctx, env)
	return httptest.NewServer(s.Handler())
}

func getWebsocketClient(server *httptest.Server) *websocket.Conn {
	serverAddr := server.Listener.Addr().String()
	wsurl := fmt.Sprintf("ws://%s%s", serverAddr, "/ws")
	wsc, _, err := websocket.DefaultDialer.Dial(wsurl, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	return wsc
}

func startOperator(t *testing.T, s *httptest.Server, ri daemon.RunInstruction) daemon.RunState {
	t.Helper()
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

func getResponse(server *httptest.Server, method string, url string, body io.Reader) *http.Response {
	request, _ := http.NewRequest(method, server.URL+url, body)
	c := server.Client()
	response, _ := c.Do(request)
	return response
}

func TestServer_Start_Operator_Push_Input_Read_Websocket_Output(t *testing.T) {
	server := getTestServer()
	wsc := getWebsocketClient(server)
	defer wsc.Close()
	defer server.Close()

	id, _ := uuid.Parse("3ceccd71-0ea5-4aeb-957a-4dff1a419071")
	data := daemon.RunInstruction{Id: id,
		Stream: false,
		Props:  core.Properties{},
		Gens:   core.Generics{},
	}

	instance := startOperator(t, server, data)
	body, _ := json.Marshal(map[string]interface{}{"input": "test"})
	response := getResponse(server, "POST", instance.URL, bytes.NewBuffer(body))
	assert.Equal(t, 200, response.StatusCode)

	_, m, _ := wsc.ReadMessage()
	type message struct {
		Topic string
		Data  interface{}
	}

	var out message
	json.Unmarshal(m, &out)
	assert.Equal(t, out.Topic, "Port")
	assert.Equal(t, out.Data, map[string]interface{}{"Data": "test", "Handle": instance.Handle, "IsBOS": false, "IsEOS": false, "Port": map[string]interface{}{}})

}

func TestServer_List_Running_Instances(t *testing.T) {
	server := getTestServer()
	id := "8b62495a-e482-4a3e-8020-0ab8a350ad2d"
	uuid, _ := uuid.Parse(id)
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
	response := getResponse(server, "GET", "/instances/", nil)
	assert.Equal(t, 200, response.StatusCode)
	body, _ := ioutil.ReadAll(response.Body)
	assert.Contains(t, string(body), id)
}
