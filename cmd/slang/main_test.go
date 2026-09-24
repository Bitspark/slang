package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/capability"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/elem"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRunRejectsUnconnectedOutputsBeforeStarting(t *testing.T) {
	for _, mode := range SupportedRunModes {
		t.Run(mode, func(t *testing.T) {
			elem.Init()
			id := uuid.New()
			blueprint := core.Blueprint{
				Id: id,
				ServiceDefs: map[string]*core.ServiceDef{core.MAIN_SERVICE: {
					In: core.TypeDef{Type: "number"},
					Out: core.TypeDef{Type: "map", Map: core.TypeDefMap{
						"a": {Type: "number"}, "b": {Type: "number"},
					}},
				}},
			}
			o, err := api.BuildOperator(&core.SlangBundle{
				Main: id, Blueprints: map[uuid.UUID]core.Blueprint{id: blueprint},
			})
			require.NoError(t, err)
			require.EqualError(t, run(o, mode, "invalid-bind-address"),
				"unconnected output ports: )a, )b")
		})
	}
}

func TestRunRejectsMissingMainService(t *testing.T) {
	o, err := core.NewOperator("empty", nil, nil, nil, nil, core.Blueprint{})
	require.NoError(t, err)
	require.EqualError(t, run(o, "process", ""), "blueprint has no main service")
}

func TestSafeModeFromEnv(t *testing.T) {
	for value, want := range map[string]bool{"": false, "false": false, "0": false, "true": true, "1": true} {
		got, err := safeModeFromEnv(value)
		require.NoError(t, err, value)
		require.Equal(t, want, got, value)
	}
	_, err := safeModeFromEnv("yes")
	require.Error(t, err, "an unreadable value must not fall back to unsafe mode")
}

// shortSocketPath avoids the length limit on unix socket paths.
func shortSocketPath(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "cli")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })
	return filepath.Join(dir, "s.sock")
}

func TestCapabilitiesFromEnvDefaultsToThisMachine(t *testing.T) {
	require.Nil(t, capabilitiesFromEnv("").HTTP.Transport)
}

func TestCapabilitiesFromEnvSendsHTTPThroughTheCapabilityHost(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "reached")
	}))
	defer target.Close()
	var hosted int64
	socket := shortSocketPath(t)
	host, err := capability.ServeUnix(socket, &capability.HTTPHost{Allow: func(*http.Request) error {
		atomic.AddInt64(&hosted, 1)
		return nil
	}})
	require.NoError(t, err)
	defer host.Close()

	resp, err := capabilitiesFromEnv(socket).HTTP.Get(target.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	require.Equal(t, "reached", string(body))
	require.Equal(t, int64(1), atomic.LoadInt64(&hosted), "the request must pass through the capability host")
}

func TestListenServesOnAUnixSocketReplacingAStaleOne(t *testing.T) {
	socket := shortSocketPath(t)
	stale, err := net.Listen("unix", socket)
	require.NoError(t, err)
	stale.(*net.UnixListener).SetUnlinkOnClose(false)
	stale.Close()

	listener, err := listen("unix:" + socket)
	require.NoError(t, err)
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "invoked")
	})}
	go server.Serve(listener)
	defer server.Close()

	client := &http.Client{Transport: &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", socket)
	}}}
	resp, err := client.Post("http://program/", "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	require.Equal(t, "invoked", string(body))
}

func TestListenKeepsAFileThatIsNotASocket(t *testing.T) {
	path := shortSocketPath(t)
	require.NoError(t, os.WriteFile(path, []byte("keep"), 0o600))
	_, err := listen("unix:" + path)
	require.Error(t, err)
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "keep", string(content))
}
