package elem

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/tests/assertions"
)

// Each test names the planning condition in slang-ecosystem/planning/ it is evidence for.

const netHTTPTestRequests = 20

func startNetHTTPClient(t *testing.T) *core.Operator {
	t.Helper()
	op, err := buildOperator(core.InstanceDef{Operator: netHTTPClientCfg.blueprint.Id})
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

func settledGoroutines() int {
	count := 0
	for i := 0; i < 20; i++ {
		runtime.GC()
		time.Sleep(50 * time.Millisecond)
		count = runtime.NumGoroutine()
	}
	return count
}

// Condition slang.http-client.reuses-connections.
func Test_NetHTTPClient__ReusesConnections(t *testing.T) {
	a := assertions.New(t)
	url, opened := connectionCountingServer(t, largeBody)
	op := startNetHTTPClient(t)
	for i := 0; i < netHTTPTestRequests; i++ {
		pushGet(op, url)
		op.Main().Out().Pull()
	}
	a.Equal(int64(1), atomic.LoadInt64(opened), "requests should share one connection")
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
	a := assertions.New(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "1000")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "truncated")
		conn, _, _ := w.(http.Hijacker).Hijack()
		conn.Close()
	}))
	defer srv.Close()
	op := startNetHTTPClient(t)
	pushGet(op, srv.URL)
	a.Nil(op.Main().Out().Pull().(map[string]interface{})["status"], "a truncated body should produce no response")
	baseline := settledGoroutines()
	for i := 0; i < netHTTPTestRequests; i++ {
		pushGet(op, srv.URL)
		op.Main().Out().Pull()
	}
	a.LessOrEqual(settledGoroutines(), baseline+2, "failed reads should not leave goroutines behind")
}

// overrideNetHTTPTimeout shortens or lengthens the operator's timeout for one test.
func overrideNetHTTPTimeout(t *testing.T, timeout time.Duration) {
	atomic.StoreInt64(&netHTTPClientTimeoutOverride, int64(timeout))
	t.Cleanup(func() { atomic.StoreInt64(&netHTTPClientTimeoutOverride, 0) })
}

// Condition slang.http-client.bounded-stall.
func Test_NetHTTPClient__GivesUpOnStalledResponse(t *testing.T) {
	a := assertions.New(t)
	overrideNetHTTPTimeout(t, 200*time.Millisecond)
	s := newStallingServer(t)
	op := startNetHTTPClient(t)
	pushGet(op, s.url)
	<-s.arrived
	select {
	case <-s.gaveUp:
	case <-time.After(5 * time.Second):
		t.Fatal("the operator kept waiting on a stalled response past its timeout")
	}
	a.Nil(op.Main().Out().Pull().(map[string]interface{})["status"], "a timed-out request should produce no response")
}

// The default must leave the program time to handle a failed request before the
// hosted runtime ends the whole invocation at 15 seconds.
func Test_NetHTTPClient__DefaultTimeoutIsBelowHostedLimit(t *testing.T) {
	a := assertions.New(t)
	a.Less(netHTTPTimeout(), 15*time.Second)
}

// Condition slang.http-client.stop-abandons-request.
func Test_NetHTTPClient__StopCancelsRequest(t *testing.T) {
	overrideNetHTTPTimeout(t, time.Minute) // only stopping may end this request
	s := newStallingServer(t)
	op := startNetHTTPClient(t)
	pushGet(op, s.url)
	<-s.arrived
	op.Stop()
	select {
	case <-s.gaveUp:
	case <-time.After(5 * time.Second):
		t.Fatal("stopping the operator left its request running")
	}
}
