// Package capability serves host resources to a program from outside its
// process, so that a host can confine the program and still grant what it may use.
//
// # HTTP
//
// A program's runner sends each request to the capability host as
//
//	POST /v1/http
//	Content-Type: message/http
//
// whose body is the request in HTTP/1.1 wire format with an absolute URL. The host
// checks the request against its policy, performs it, reads the whole response,
// and answers 200 with the response in the same format. Any other status means the
// host did not deliver a response: 403 when policy refused the request, 502 when
// performing it failed, 400 when the envelope was malformed. The body of such an
// answer is a plain-text reason.
//
// Cancelling the runner's request closes its connection to the host, which
// cancels the request the host is making.
package capability

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"strings"
)

// HTTPPath is where a capability host serves HTTP requests.
const HTTPPath = "/v1/http"

const messageType = "message/http"

// HTTPHost performs HTTP requests on behalf of a program.
type HTTPHost struct {
	// Client performs the requests. A nil Client uses a client without its own
	// timeout, leaving the bound to the runner.
	Client *http.Client
	// Allow decides whether a request may be performed. A nil Allow permits every
	// request.
	Allow func(*http.Request) error
}

func (h *HTTPHost) client() *http.Client {
	if h.Client != nil {
		return h.Client
	}
	return http.DefaultClient
}

func (h *HTTPHost) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != HTTPPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "use POST", http.StatusMethodNotAllowed)
		return
	}
	if mediaType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type")); mediaType != messageType {
		http.Error(w, "the body must be "+messageType, http.StatusUnsupportedMediaType)
		return
	}

	req, err := http.ReadRequest(bufio.NewReader(r.Body))
	if err != nil {
		http.Error(w, "unreadable request: "+err.Error(), http.StatusBadRequest)
		return
	}
	if !req.URL.IsAbs() {
		http.Error(w, "the request needs an absolute URL", http.StatusBadRequest)
		return
	}
	req.RequestURI = ""
	req = req.WithContext(r.Context())

	if h.Allow != nil {
		if err := h.Allow(req); err != nil {
			http.Error(w, "refused: "+err.Error(), http.StatusForbidden)
			return
		}
	}

	resp, err := h.client().Do(req)
	if err != nil {
		http.Error(w, "request failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "response failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	delivered := &http.Response{
		Status:        resp.Status,
		StatusCode:    resp.StatusCode,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        resp.Header.Clone(),
		Body:          io.NopCloser(bytes.NewReader(body)),
		ContentLength: int64(len(body)),
	}
	var wire bytes.Buffer
	if err := delivered.Write(&wire); err != nil {
		http.Error(w, "response could not be encoded: "+err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", messageType)
	w.WriteHeader(http.StatusOK)
	w.Write(wire.Bytes())
}

// httpTransport sends requests to a capability host over a unix socket.
type httpTransport struct {
	envelopes *http.Transport
}

// NewHTTPTransport returns a RoundTripper that performs each request through the
// capability host listening on the unix socket at path. Timeouts and cancellation
// come from the request's context, as with any RoundTripper.
func NewHTTPTransport(socket string) http.RoundTripper {
	return &httpTransport{envelopes: &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var dialer net.Dialer
			return dialer.DialContext(ctx, "unix", socket)
		},
		DisableCompression: true,
	}}
}

func (t *httpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var wire bytes.Buffer
	if err := req.WriteProxy(&wire); err != nil {
		return nil, err
	}
	if req.Body != nil {
		req.Body.Close()
	}

	envelope, err := http.NewRequestWithContext(req.Context(), http.MethodPost, "http://capability-host"+HTTPPath, &wire)
	if err != nil {
		return nil, err
	}
	envelope.Header.Set("Content-Type", messageType)
	answer, err := t.envelopes.RoundTrip(envelope)
	if err != nil {
		return nil, err
	}
	if answer.StatusCode != http.StatusOK {
		defer answer.Body.Close()
		reason, _ := io.ReadAll(io.LimitReader(answer.Body, 4096))
		return nil, fmt.Errorf("capability host: %s: %s", answer.Status, strings.TrimSpace(string(reason)))
	}

	resp, err := http.ReadResponse(bufio.NewReader(answer.Body), req)
	if err != nil {
		answer.Body.Close()
		return nil, err
	}
	resp.Body = &deliveredBody{ReadCloser: resp.Body, envelope: answer.Body}
	return resp, nil
}

// deliveredBody closes the envelope together with the response it carried,
// draining it first so the connection to the host can be reused.
type deliveredBody struct {
	io.ReadCloser
	envelope io.ReadCloser
}

func (b *deliveredBody) Close() error {
	err := b.ReadCloser.Close()
	io.Copy(io.Discard, b.envelope)
	if closeErr := b.envelope.Close(); err == nil {
		err = closeErr
	}
	return err
}

// ServeUnix serves handler on a new unix socket at path until the returned server
// is closed.
func ServeUnix(socket string, handler http.Handler) (*http.Server, error) {
	listener, err := net.Listen("unix", socket)
	if err != nil {
		return nil, err
	}
	server := &http.Server{Handler: handler}
	go server.Serve(listener)
	return server, nil
}
