package capability

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// socketPath returns a short path for a unix socket; long temporary paths exceed
// the platform limit on socket names.
func socketPath(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "cap")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })
	return filepath.Join(dir, "h.sock")
}

// startHost serves host on a new socket and returns a client that reaches the
// world only through it, plus the number of connections the host accepted.
func startHost(t *testing.T, host *HTTPHost) (*http.Client, *int64) {
	t.Helper()
	socket := socketPath(t)
	listener, err := net.Listen("unix", socket)
	require.NoError(t, err)
	var accepted int64
	server := &http.Server{Handler: host, ConnState: func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			atomic.AddInt64(&accepted, 1)
		}
	}}
	go server.Serve(listener)
	t.Cleanup(func() { server.Close() })
	return &http.Client{Transport: NewHTTPTransport(socket), Timeout: 10 * time.Second}, &accepted
}

func TestHTTPRoundTripPreservesTheExchange(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("X-Seen", r.Method+" "+r.URL.RequestURI()+" "+r.Header.Get("X-Token")+" "+string(body))
		w.Header().Add("X-Repeated", "a")
		w.Header().Add("X-Repeated", "b")
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, "made")
	}))
	defer target.Close()
	client, _ := startHost(t, &HTTPHost{})

	req, err := http.NewRequest(http.MethodPost, target.URL+"/things?kind=a%20b", strings.NewReader("payload"))
	require.NoError(t, err)
	req.Header.Set("X-Token", "secret")
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.Equal(t, "made", string(body))
	require.Equal(t, "POST /things?kind=a%20b secret payload", resp.Header.Get("X-Seen"))
	require.Equal(t, []string{"a", "b"}, resp.Header.Values("X-Repeated"))
}

// Writing a request in origin form would lose its scheme and send HTTPS as HTTP.
func TestHTTPRoundTripKeepsHTTPS(t *testing.T) {
	target := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "secure")
	}))
	defer target.Close()
	client, _ := startHost(t, &HTTPHost{Client: target.Client()})

	resp, err := client.Get(target.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	require.Equal(t, "secure", string(body))
}

func TestHTTPErrorStatusesPassThrough(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "gone", http.StatusNotFound)
	}))
	defer target.Close()
	client, _ := startHost(t, &HTTPHost{})

	resp, err := client.Get(target.URL)
	require.NoError(t, err, "a 404 is a delivered response, not a failure")
	defer resp.Body.Close()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHTTPHostRefusesWhatPolicyForbids(t *testing.T) {
	var reached int64
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&reached, 1)
	}))
	defer target.Close()
	client, _ := startHost(t, &HTTPHost{Allow: func(r *http.Request) error {
		return errors.New("no destination allowed")
	}})

	_, err := client.Get(target.URL)
	require.ErrorContains(t, err, "403")
	require.ErrorContains(t, err, "no destination allowed")
	require.Zero(t, atomic.LoadInt64(&reached), "a refused request must not reach its target")
}

func TestHTTPHostRejectsMalformedEnvelopes(t *testing.T) {
	host := &HTTPHost{}
	for name, envelope := range map[string]struct {
		method, path, contentType, body string
		status                          int
	}{
		"wrong path":         {"POST", "/v2/http", messageType, "", http.StatusNotFound},
		"wrong method":       {"GET", HTTPPath, messageType, "", http.StatusMethodNotAllowed},
		"wrong content type": {"POST", HTTPPath, "application/json", "{}", http.StatusUnsupportedMediaType},
		"not a request":      {"POST", HTTPPath, messageType, "hello", http.StatusBadRequest},
		"relative URL":       {"POST", HTTPPath, messageType, "GET /x HTTP/1.1\r\nHost: a\r\n\r\n", http.StatusBadRequest},
	} {
		t.Run(name, func(t *testing.T) {
			r := httptest.NewRequest(envelope.method, envelope.path, strings.NewReader(envelope.body))
			r.Header.Set("Content-Type", envelope.contentType)
			w := httptest.NewRecorder()
			host.ServeHTTP(w, r)
			require.Equal(t, envelope.status, w.Code)
		})
	}
}

func TestHTTPCancellationReachesTheTarget(t *testing.T) {
	arrived, gaveUp, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(arrived)
		select {
		case <-r.Context().Done():
			close(gaveUp)
		case <-release:
		}
	}))
	defer target.Close()
	defer close(release)
	client, _ := startHost(t, &HTTPHost{})

	ctx, cancel := context.WithCancel(context.Background())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.URL, nil)
	require.NoError(t, err)
	go client.Do(req)
	<-arrived
	cancel()
	select {
	case <-gaveUp:
	case <-time.After(5 * time.Second):
		t.Fatal("cancelling the runner's request left the host's request running")
	}
}

func TestHTTPTransportReusesItsConnectionToTheHost(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, strings.Repeat("x", 64<<10))
	}))
	defer target.Close()
	client, accepted := startHost(t, &HTTPHost{})

	for i := 0; i < 20; i++ {
		resp, err := client.Get(target.URL)
		require.NoError(t, err)
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
	require.Equal(t, int64(1), atomic.LoadInt64(accepted), "requests should share one connection to the host")
}
