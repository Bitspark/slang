package elem

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Bitspark/slang/pkg/capability"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
)

// Each test names the planning condition in slang-ecosystem/planning/ it is evidence for.

const netHTTPTestRequests = 20

func startNetHTTPClient(t *testing.T) *core.Operator {
	t.Helper()
	return startNetHTTPClientWith(t, LocalCapabilities())
}

func startNetHTTPClientWith(t *testing.T, caps Capabilities) *core.Operator {
	t.Helper()
	op, err := buildOperatorWith(core.InstanceDef{Operator: netHTTPClientCfg.blueprint.Id}, caps)
	if err != nil {
		t.Fatal(err)
	}
	op.Main().Out().Bufferize()
	go op.Start()
	return op
}

func pushGet(op *core.Operator, url string) {
	op.Main().In().Push(map[string]interface{}{
		"method": "GET", "url": url, "headers": []interface{}{}, "body": core.Binary{},
	})
}

// connectionCountingServer records every TCP connection a client opens to it.
func connectionCountingServer(t *testing.T, h http.HandlerFunc) (string, *int64) {
	var opened int64
	s := httptest.NewUnstartedServer(h)
	s.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			atomic.AddInt64(&opened, 1)
		}
	}
	s.Start()
	t.Cleanup(s.Close)
	return s.URL, &opened
}

func largeBody(w http.ResponseWriter, _ *http.Request) {
	io.WriteString(w, strings.Repeat("x", 64<<10))
}

type stallingServer struct {
	url     string
	arrived chan struct{} // closed once a request is being answered
	gaveUp  chan struct{} // closed once the client abandons the response
}

// waitForRequest fails the test if no request reaches the server in time, rather
// than letting it wait for the test binary's timeout.
func (s *stallingServer) waitForRequest(t *testing.T) {
	t.Helper()
	select {
	case <-s.arrived:
	case <-time.After(5 * time.Second):
		t.Fatal("the request never reached the server")
	}
}

// newStallingServer sends headers and part of a body, then waits until the
// client disconnects or the test ends.
func newStallingServer(t *testing.T) *stallingServer {
	release := make(chan struct{})
	s := &stallingServer{arrived: make(chan struct{}), gaveUp: make(chan struct{})}
	var arrive, leave sync.Once
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "partial")
		w.(http.Flusher).Flush()
		arrive.Do(func() { close(s.arrived) })
		select {
		case <-r.Context().Done():
			leave.Do(func() { close(s.gaveUp) })
		case <-release:
		}
	}))
	t.Cleanup(func() {
		close(release)
		srv.Close()
	})
	s.url = srv.URL
	return s
}

// bodyCountingTransport records every response body it hands out and every one
// that is closed again.
type bodyCountingTransport struct{ opened, closed int64 }

type countedBody struct {
	io.ReadCloser
	closed *int64
	once   sync.Once
}

func (b *countedBody) Close() error {
	b.once.Do(func() { atomic.AddInt64(b.closed, 1) })
	return b.ReadCloser.Close()
}

func (c *bodyCountingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := http.DefaultTransport.RoundTrip(r)
	if err == nil {
		atomic.AddInt64(&c.opened, 1)
		resp.Body = &countedBody{ReadCloser: resp.Body, closed: &c.closed}
	}
	return resp, err
}

// httpImplementation provides the HTTP capability either in process or through a
// capability host on a unix socket. The operator must behave the same with both.
type httpImplementation struct {
	name string
	// caps returns capabilities whose requests give up after timeout. outbound, if
	// not nil, performs the requests that reach the target server.
	caps func(t *testing.T, timeout time.Duration, outbound http.RoundTripper) Capabilities
}

var httpImplementations = []httpImplementation{
	{"in-process", func(t *testing.T, timeout time.Duration, outbound http.RoundTripper) Capabilities {
		return Capabilities{HTTP: &http.Client{Timeout: timeout, Transport: outbound}}
	}},
	{"capability-host", func(t *testing.T, timeout time.Duration, outbound http.RoundTripper) Capabilities {
		dir, err := os.MkdirTemp("", "cap") // short: socket paths have a length limit
		if err != nil {
			t.Fatal(err)
		}
		socket := filepath.Join(dir, "h.sock")
		host, err := capability.ServeUnix(socket, &capability.HTTPHost{Client: &http.Client{Transport: outbound}})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			host.Close()
			os.RemoveAll(dir)
		})
		return Capabilities{HTTP: &http.Client{Timeout: timeout, Transport: capability.NewHTTPTransport(socket)}}
	}},
}

// forEachHTTPImplementation runs test once per implementation of the HTTP capability.
func forEachHTTPImplementation(t *testing.T, test func(t *testing.T, impl httpImplementation)) {
	for _, impl := range httpImplementations {
		impl := impl
		t.Run(impl.name, func(t *testing.T) { test(t, impl) })
	}
}

func Test_NetHTTPClient__DeliversResponses(t *testing.T) {
	forEachHTTPImplementation(t, func(t *testing.T, impl httpImplementation) {
		a := assertions.New(t)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Answer", "42")
			w.WriteHeader(http.StatusCreated)
			io.WriteString(w, "made")
		}))
		defer srv.Close()
		op := startNetHTTPClientWith(t, impl.caps(t, 10*time.Second, nil))
		pushGet(op, srv.URL)
		resp := op.Main().Out().Pull().(map[string]interface{})
		a.Equal(201.0, resp["status"])
		a.Equal(core.Binary("made"), resp["body"])
		a.Contains(resp["headers"], map[string]interface{}{"key": "X-Answer", "value": "42"})
	})
}

// Condition slang.http-client.reuses-connections.
func Test_NetHTTPClient__ReusesConnections(t *testing.T) {
	forEachHTTPImplementation(t, func(t *testing.T, impl httpImplementation) {
		a := assertions.New(t)
		url, opened := connectionCountingServer(t, largeBody)
		op := startNetHTTPClientWith(t, impl.caps(t, 10*time.Second, nil))
		for i := 0; i < netHTTPTestRequests; i++ {
			pushGet(op, url)
			op.Main().Out().Pull()
		}
		a.Equal(int64(1), atomic.LoadInt64(opened), "requests should share one connection")
	})
}

// Control for the test above: a client that genuinely leaks must be detected.
func Test_NetHTTPClient__HarnessDetectsLeakedConnections(t *testing.T) {
	a := assertions.New(t)
	url, opened := connectionCountingServer(t, largeBody)
	for i := 0; i < netHTTPTestRequests; i++ {
		resp, err := http.Get(url)
		a.NoError(err)
		resp.Body.Read(make([]byte, 16)) // read partially and never close
	}
	a.Equal(int64(netHTTPTestRequests), atomic.LoadInt64(opened), "each leaked response should hold its connection")
}

// Condition slang.http-client.recovers-from-read-errors.
func Test_NetHTTPClient__ReleasesConnectionsAfterReadErrors(t *testing.T) {
	forEachHTTPImplementation(t, func(t *testing.T, impl httpImplementation) {
		a := assertions.New(t)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Length", "1000")
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, "truncated")
			conn, _, _ := w.(http.Hijacker).Hijack()
			conn.Close()
		}))
		defer srv.Close()
		// The counting transport makes the requests that reach the server: the
		// operator's own in process, the capability host's when remote.
		bodies := &bodyCountingTransport{}
		op := startNetHTTPClientWith(t, impl.caps(t, 10*time.Second, bodies))
		for i := 0; i < netHTTPTestRequests; i++ {
			pushGet(op, srv.URL)
			a.Nil(op.Main().Out().Pull().(map[string]interface{})["status"], "a truncated body should produce no response")
		}
		a.Equal(int64(netHTTPTestRequests), atomic.LoadInt64(&bodies.opened), "every request should reach the server")
		a.Equal(atomic.LoadInt64(&bodies.opened), atomic.LoadInt64(&bodies.closed), "every response body should be closed after a failed read")
	})
}

// Condition slang.http-client.bounded-stall.
func Test_NetHTTPClient__GivesUpOnStalledResponse(t *testing.T) {
	forEachHTTPImplementation(t, func(t *testing.T, impl httpImplementation) {
		a := assertions.New(t)
		s := newStallingServer(t)
		op := startNetHTTPClientWith(t, impl.caps(t, 200*time.Millisecond, nil))
		pushGet(op, s.url)
		s.waitForRequest(t)
		select {
		case <-s.gaveUp:
		case <-time.After(5 * time.Second):
			t.Fatal("the operator kept waiting on a stalled response past its timeout")
		}
		a.Nil(op.Main().Out().Pull().(map[string]interface{})["status"], "a timed-out request should produce no response")
	})
}

// The default must leave the program time to handle a failed request before the
// hosted runtime ends the whole invocation at 15 seconds.
func Test_NetHTTPClient__DefaultTimeoutIsBelowHostedLimit(t *testing.T) {
	a := assertions.New(t)
	a.Less(LocalCapabilities().HTTP.Timeout, 15*time.Second)
}

// Condition slang.http-client.stop-abandons-request.
func Test_NetHTTPClient__StopCancelsRequest(t *testing.T) {
	forEachHTTPImplementation(t, func(t *testing.T, impl httpImplementation) {
		s := newStallingServer(t)
		op := startNetHTTPClientWith(t, impl.caps(t, time.Minute, nil)) // only stopping may end this request
		pushGet(op, s.url)
		s.waitForRequest(t)
		op.Stop()
		select {
		case <-s.gaveUp:
		case <-time.After(5 * time.Second):
			t.Fatal("stopping the operator left its request running")
		}
	})
}

func Test_NetHTTPClient__UnavailableWithoutHTTPCapability(t *testing.T) {
	a := assertions.New(t)
	_, err := buildOperatorWith(core.InstanceDef{Operator: netHTTPClientCfg.blueprint.Id}, Capabilities{})
	a.ErrorContains(err, "does not provide HTTP")
}
