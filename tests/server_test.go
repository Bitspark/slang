package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/Bitspark/slang/pkg/daemon"
	"github.com/Bitspark/slang/pkg/env"
	"github.com/Bitspark/slang/pkg/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	dir := t.TempDir()
	for _, key := range []string{"SLANG_DIR", "SLANG_LIB_REPO_PATH", "SLANG_LIB", "SLANG_UI"} {
		t.Setenv(key, filepath.Join(dir, key))
	}
	environment := env.New("localhost", 8000)
	st := storage.NewStorage().AddBackend(storage.NewReadOnlyFileSystem("../fixtures"))
	ctx := daemon.SetStorage(context.Background(), st)
	s := daemon.NewServer(&ctx, environment, nil)
	server := httptest.NewServer(s.Handler())
	server.Client().Timeout = 5 * time.Second
	t.Cleanup(server.Close)
	return server
}

func startOperator(t *testing.T, server *httptest.Server) daemon.ResponseRunOp {
	t.Helper()
	data := daemon.RequestRunOp{Blueprint: uuid.MustParse("3ceccd71-0ea5-4aeb-957a-4dff1a419071")}
	body, err := json.Marshal(data)
	require.NoError(t, err)
	response, err := server.Client().Post(server.URL+"/run/", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer response.Body.Close()
	var out daemon.ResponseRunOp
	require.NoError(t, json.NewDecoder(response.Body).Decode(&out))
	require.Equal(t, http.StatusOK, response.StatusCode, "%+v", out.Error)
	require.Equal(t, "ok", out.Status)
	require.NotNil(t, out.Object)
	t.Cleanup(func() {
		request, err := http.NewRequest(http.MethodDelete, server.URL+out.URL(), nil)
		require.NoError(t, err)
		response, err := server.Client().Do(request)
		require.NoError(t, err)
		defer response.Body.Close()
		require.Equal(t, http.StatusNoContent, response.StatusCode)
	})
	return out
}

func TestServerStartOperatorReturnsHTTPOutput(t *testing.T) {
	server := newTestServer(t)
	instance := startOperator(t, server)
	for _, input := range []string{"first", "second", "third"} {
		body, err := json.Marshal(map[string]interface{}{"input": input})
		require.NoError(t, err)
		response, err := server.Client().Post(server.URL+instance.URL(), "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		var out map[string]interface{}
		err = json.NewDecoder(response.Body).Decode(&out)
		response.Body.Close()
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)
		require.Equal(t, map[string]interface{}{"output": input}, out)
	}
}

func TestServerListsRunningInstances(t *testing.T) {
	server := newTestServer(t)
	instance := startOperator(t, server)
	response, err := server.Client().Get(server.URL + "/run/")
	require.NoError(t, err)
	defer response.Body.Close()
	var out struct {
		Status  string `json:"status"`
		Objects []struct {
			Handle string `json:"handle"`
		} `json:"objects"`
	}
	require.NoError(t, json.NewDecoder(response.Body).Decode(&out))
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, "ok", out.Status)
	require.Len(t, out.Objects, 1)
	require.Equal(t, instance.Handle(), out.Objects[0].Handle)
}
