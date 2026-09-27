package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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

// Each invocation must receive the output of its own input, however many arrive
// at once: the running operator is shared by all of them.
func TestHttpPostPairsEachInvocationWithItsOwnOutput(t *testing.T) {
	elem.Init()
	id := uuid.New()
	echo := core.Blueprint{
		Id: id,
		ServiceDefs: map[string]*core.ServiceDef{core.MAIN_SERVICE: {
			In:  core.TypeDef{Type: "number"},
			Out: core.TypeDef{Type: "number"},
		}},
		Connections: map[string][]string{"(": {")"}},
	}
	o, err := api.BuildOperator(&core.SlangBundle{
		Main: id, Blueprints: map[uuid.UUID]core.Blueprint{id: echo},
	})
	require.NoError(t, err)
	server := httptest.NewServer(httpPostHandler(o))
	defer server.Close()
	o.Main().Out().Bufferize()
	o.Start()
	defer o.Stop()

	const callers, calls = 32, 25
	client := &http.Client{Transport: &http.Transport{MaxIdleConnsPerHost: callers}}
	var mismatched atomic.Int32
	var wg sync.WaitGroup
	for c := 0; c < callers; c++ {
		wg.Add(1)
		go func(c int) {
			defer wg.Done()
			for n := 0; n < calls; n++ {
				want := strconv.Itoa(c*calls + n)
				resp, err := client.Post(server.URL, "application/json", strings.NewReader(want))
				if err != nil {
					t.Error(err)
					return
				}
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				if strings.TrimSpace(string(body)) != want {
					mismatched.Add(1)
				}
			}
		}(c)
	}
	wg.Wait()
	require.Zero(t, mismatched.Load(), "invocations that received another invocation's output")
}

// goneCaller is a response whose caller has disconnected: every write fails.
type goneCaller struct{ header http.Header }

func (g *goneCaller) Header() http.Header       { return g.header }
func (g *goneCaller) WriteHeader(int)           {}
func (g *goneCaller) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

// A caller that disconnects before its response is written must not end the
// program for everyone else; before, this exited the process.
func TestRespondingToAGoneCallerKeepsTheProgramRunning(t *testing.T) {
	responseWithOk(&goneCaller{header: http.Header{}}, map[string]interface{}{"a": 1})
	responseWithError(&goneCaller{header: http.Header{}}, io.EOF, http.StatusBadRequest)
}
