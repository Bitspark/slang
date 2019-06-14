package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/stretchr/testify/assert"
)

func newTestServer(deployer Deployer) *httptest.Server {
	return httptest.NewServer(newRouter(deployer))
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
	slangFile := getSlangFile()
	_, err := dh.Deploy(slangFile, Process)
	assert.NoError(t, err)

	instances := dh.List()
	assert.Len(t, instances, 1)
	assert.Equal(t, slangFile.Main, instances[0].Operator)
}

func TestServer_DeployInstance(t *testing.T) {
	server := newTestServer(newTestDeployer())
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

func TestServer_GetInstance(t *testing.T) {
	dh := newTestDeployer()
	slangFile := getSlangFile()
	instance, _ := dh.Deploy(slangFile, Process)
	/****/

	server := newTestServer(dh)
	defer server.Close()

	var data interface{}
	r := getResponse(t, server, "GET", fmt.Sprint("/api/v1/instances/", instance.Instance), data)
	assert.Equal(t, http.StatusOK, r.StatusCode)

	var instance2 InstanceDef

	body, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	err = json.Unmarshal(body, &instance2)
	assert.NoError(t, err)

	assert.Equal(t, instance.Instance, instance2.Instance)
	assert.Equal(t, instance.Operator, instance2.Operator)
	assert.Equal(t, instance.Mode, instance2.Mode)
}

func TestServer_ListInstance(t *testing.T) {
	dh := newTestDeployer()
	slangFile := getSlangFile()
	dh.Deploy(slangFile, Process)
	dh.Deploy(slangFile, Process)
	/****/

	server := newTestServer(dh)
	defer server.Close()

	var data interface{}
	r := getResponse(t, server, "GET", "/api/v1/instances", data)
	assert.Equal(t, http.StatusOK, r.StatusCode)

	var instances []InstanceDef

	body, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	err = json.Unmarshal(body, &instances)
	assert.NoError(t, err)

	assert.Len(t, instances, 2)
}
