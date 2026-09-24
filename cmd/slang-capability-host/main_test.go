package main

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Bitspark/slang/pkg/capability"
	"github.com/stretchr/testify/require"
)

func TestRefusedNetworksAlwaysIncludePrivateRanges(t *testing.T) {
	refused, err := refusedNetworks("")
	require.NoError(t, err)
	require.Len(t, refused, len(capability.PrivateNetworks))

	refused, err = refusedNetworks("198.51.100.7/32,203.0.113.9/32")
	require.NoError(t, err)
	require.Len(t, refused, len(capability.PrivateNetworks)+2)

	_, err = refusedNetworks("198.51.100.7")
	require.Error(t, err, "an address without a prefix length is not a network")
}

func TestCapabilityHostRefusesPrivateTargets(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("a private target must not be reached")
	}))
	defer target.Close()
	refused, err := refusedNetworks("")
	require.NoError(t, err)
	host := httptest.NewServer(capabilityHost(refused, time.Second))
	defer host.Close()

	request := "GET " + target.URL + "/ HTTP/1.1\r\nHost: " + strings.TrimPrefix(target.URL, "http://") + "\r\n\r\n"
	resp, err := http.Post(host.URL+capability.HTTPPath, "message/http", strings.NewReader(request))
	require.NoError(t, err)
	defer resp.Body.Close()
	reason, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusBadGateway, resp.StatusCode)
	require.Contains(t, string(reason), "not reachable from programs")
}

func TestInvocationProxyPassesInvocationsToTheProgram(t *testing.T) {
	dir, err := os.MkdirTemp("", "inv")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	socket := filepath.Join(dir, "invoke.sock")
	listener, err := net.Listen("unix", socket)
	require.NoError(t, err)
	program := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusAccepted)
		io.WriteString(w, r.Method+" "+r.URL.Path+" "+string(body))
	})}
	go program.Serve(listener)
	defer program.Close()

	gateway := httptest.NewServer(invocationProxy(socket))
	defer gateway.Close()
	resp, err := http.Post(gateway.URL+"/", "application/json", strings.NewReader(`{"a":1}`))
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusAccepted, resp.StatusCode)
	require.Equal(t, `POST / {"a":1}`, string(body))
}
