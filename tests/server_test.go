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

type message struct {
	Topic string
	Data  interface{}
}

func newTestServer() *httptest.Server {
	env := env.New("localhost", 8000)
	// This storage allows us to load prebuilt operators for tests,
	// it's also useful to have regressions there.
	st := storage.NewStorage().
		AddBackend(storage.NewReadOnlyFileSystem("../fixtures"))

	ctx := daemon.SetStorage(context.Background(), st)
	s := daemon.NewServer(&ctx, env)
	return httptest.NewServer(s.Handler())
}

func newWebsocketClient(t *testing.T, server *httptest.Server) *websocket.Conn {
	serverAddr := server.Listener.Addr().String()
	wsurl := fmt.Sprintf("ws://%s%s", serverAddr, "/ws")
	wsc, _, err := websocket.DefaultDialer.Dial(wsurl, nil)
	if err != nil {
		t.Fatal(err)
	}
	return wsc
}

func readOneMessage(t *testing.T, wsc *websocket.Conn) message {
	var out message
	var err error

	// This reads exactly one message that was send via the websocket
	_, m, err := wsc.ReadMessage()

	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(m, &out)
	if err != nil {
		t.Fatal(err)
	}

	return out
}

func startOperator(t *testing.T, s *httptest.Server, ri daemon.RunInstruction) daemon.RunState {
	t.Helper() // this allows fail asserts to actually mark the calling function as a failure
	var out daemon.RunState
	var err error

	body, _ := json.Marshal(&ri)
	request, _ := http.NewRequest("POST", s.URL+"/run/", bytes.NewReader(body))
	response, err := s.Client().Do(request)
	if err != nil {
		t.Fatal(err)
	}
	body, _ = ioutil.ReadAll(response.Body)
	err = json.Unmarshal(body, &out)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 200, response.StatusCode)
	assert.Equal(t, "success", out.Status)
	return out
}

func getResponse(t *testing.T, server *httptest.Server, method string, url string, body io.Reader) *http.Response {
	request, _ := http.NewRequest(method, server.URL+url, body)
	c := server.Client()
	response, err := c.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func TestServer_Start_Operator_Push_Input_Read_Websocket_Output(t *testing.T) {
	server := newTestServer()
	wsc := newWebsocketClient(t, server)
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
	response := getResponse(t, server, "POST", instance.URL, bytes.NewBuffer(body))
	assert.Equal(t, 200, response.StatusCode)
	out := readOneMessage(t, wsc)
	assert.Equal(t, out.Topic, "Port")
	assert.Equal(t, out.Data, map[string]interface{}{"Data": "test", "Handle": instance.Handle, "IsBOS": false, "IsEOS": false, "Port": map[string]interface{}{}})

}

func TestServer_List_Running_Instances(t *testing.T) {
	server := newTestServer()
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
	response := getResponse(t, server, "GET", "/instances/", nil)
	assert.Equal(t, 200, response.StatusCode)
	body, _ := ioutil.ReadAll(response.Body)
	assert.Contains(t, string(body), id)
}
