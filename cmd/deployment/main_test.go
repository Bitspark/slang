package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Bitspark/go-funk"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type testInstanceManager struct {
	list []*InstanceDef
}

func (m testInstanceManager) List() []InstanceDef {
	return funk.Map(funk.Filter(m.list, func(i *InstanceDef) bool {
		return i != nil
	}), func(i *InstanceDef) InstanceDef {
		return *i
	}).([]InstanceDef)
}

func (m *testInstanceManager) Start(instance uuid.UUID, mode T_DeploymentMode, def core.SlangFileDef) (InstanceDef, error) {
	ins := InstanceDef{instance, def.Main, Running, mode}
	m.list = append(m.list, &ins)
	return ins, nil
}
func (m testInstanceManager) Restart(instance uuid.UUID) error {
	return nil
}

func (m testInstanceManager) Stop(instance uuid.UUID) error {
	if funk.Find(m.List(), func(i *InstanceDef) bool { return i.Instance == instance }) != nil {
		return fmt.Errorf("unknown instance")
	}
	return nil
}

func (m testInstanceManager) Info(instance uuid.UUID) (InstanceDef, error) {
	if ins := funk.Find(m.List(), func(i *InstanceDef) bool { return i.Instance == instance }).(*InstanceDef); ins != nil {
		return *ins, nil
	}
	return InstanceDef{uuid.Nil, uuid.Nil, "", ""}, fmt.Errorf("unknown instance")
}

func newTestDeployer() Deployer {
	return newDeployer(&testInstanceManager{})
}

func newTestServer() *httptest.Server {
	return httptest.NewServer(newRouter(newTestDeployer()))
}

func getResponse(t *testing.T, server *httptest.Server, method string, url string, data interface{}) *http.Response {
	body, _ := json.Marshal(data)
	request, _ := http.NewRequest(method, server.URL+url, bytes.NewBuffer(body))
	c := server.Client()
	response, err := c.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func getSlangFile() core.SlangFileDef {
	var slangFile core.SlangFileDef
	if err := json.Unmarshal([]byte(`{
  "main": "1fd6b4b2-c87c-4c71-8046-2a94f9f35af5",
  "blueprints": [
    {
      "id": "1fd6b4b2-c87c-4c71-8046-2a94f9f35af5",
      "operators": {
        "now": {
          "operator": "808c7846-db9f-43ee-989b-37a08ce7e70d"
        }
      },
      "services": {
        "main": {
          "in": {
            "type": "map",
            "map": { "gen__57": { "type": "trigger" } }
          },
          "out": {
            "type": "map",
            "map": {
              "gen_second_73": { "type": "number" },
              "gen_minute_31": { "type": "number" },
              "gen_hour_68": { "type": "number" }
            }
          }
        }
      },
      "connections": {
        "gen__57(": ["(now"],
        "now)hour": [")gen_hour_68"],
        "now)minute": [")gen_minute_31"],
        "now)second": [")gen_second_73"]
      }
    },
    {
      "id": "808c7846-db9f-43ee-989b-37a08ce7e70d",
      "operators": {},
      "services": {
        "main": {
          "in": { "type": "trigger" },
          "geometry": {
            "in": { "position": 0 },
            "out": { "position": 0 }
          },
          "out": {
            "type": "map",
            "map": {
              "day": { "type": "number" },
              "hour": { "type": "number" },
              "minute": { "type": "number" },
              "month": { "type": "number" },
              "nanosecond": { "type": "number" },
              "second": { "type": "number" },
              "year": { "type": "number" }
            }
          }
        }
      },
      "delegates": {},
      "properties": {},
      "connections": {}
    }
  ]
}`), &slangFile); err != nil {
		panic(err)
	}

	return slangFile
}

func TestDeployer_Deploy(t *testing.T) {
	dh := newTestDeployer()

	assert.Empty(t, dh.List())
	_, err := dh.Deploy(getSlangFile(), Process)
	assert.NoError(t, err)
	assert.NotEmpty(t, dh.List())
}

func TestServer_Deploy(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	data := DeployInstruction{
		getSlangFile(),
		Process,
	}

	r := getResponse(t, server, "POST", "/api/v1/instances", data)
	assert.Equal(t, http.StatusOK, r.StatusCode)

	var instance InstanceDef

	body, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	err = json.Unmarshal(body, &instance)
	assert.NoError(t, err)

	assert.Equal(t, data.SlangFile.Main, instance.Operator)
	assert.Equal(t, data.Mode, instance.Mode)
}
