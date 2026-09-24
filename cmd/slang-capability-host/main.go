// Command slang-capability-host serves one execution's capabilities from beside
// its runner, which has no network of its own. The two share a socket directory:
//
//	capability.sock  HTTP requests from the program, performed here after checking
//	                 the address actually dialed against the refused networks
//	invoke.sock      the program's own HTTP service, to which invocations arriving
//	                 on -bind are passed
package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Bitspark/slang/pkg/capability"
)

func main() {
	dir := flag.String("dir", "", "Socket directory shared with the runner")
	bind := flag.String("bind", "", "Address on which to accept invocations; empty accepts none")
	refuse := flag.String("refuse", "", "Comma-separated networks refused in addition to private and reserved ranges")
	timeout := flag.Duration("timeout", 30*time.Second, "Longest a single HTTP request may take")
	flag.Parse()
	if *dir == "" {
		log.Fatal("-dir is required")
	}

	refused, err := refusedNetworks(*refuse)
	if err != nil {
		log.Fatal(err)
	}
	socket := filepath.Join(*dir, "capability.sock")
	if info, err := os.Lstat(socket); err == nil && info.Mode()&os.ModeSocket != 0 {
		os.Remove(socket) // left behind by an earlier run of this execution
	}
	host, err := capability.ServeUnix(socket, capabilityHost(refused, *timeout))
	if err != nil {
		log.Fatal(err)
	}
	defer host.Close()

	if *bind != "" {
		invocations := &http.Server{Addr: *bind, Handler: invocationProxy(filepath.Join(*dir, "invoke.sock"))}
		go func() { log.Fatal(invocations.ListenAndServe()) }()
		defer invocations.Close()
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}

// refusedNetworks adds the comma-separated networks in extra to the private and
// reserved ranges that are always refused.
func refusedNetworks(extra string) ([]*net.IPNet, error) {
	refused := append([]*net.IPNet{}, capability.PrivateNetworks...)
	if extra == "" {
		return refused, nil
	}
	more, err := capability.ParseNetworks(strings.Split(extra, ","))
	if err != nil {
		return nil, err
	}
	return append(refused, more...), nil
}

// capabilityHost performs a program's HTTP requests, refusing every connection to
// a refused network, including those a redirect leads to.
func capabilityHost(refused []*net.IPNet, timeout time.Duration) *capability.HTTPHost {
	dialer := &net.Dialer{Timeout: 5 * time.Second, Control: capability.RefuseAddresses(refused)}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = dialer.DialContext
	return &capability.HTTPHost{Client: &http.Client{Transport: transport, Timeout: timeout}}
}

// invocationProxy passes invocations to the program's HTTP service on socket.
func invocationProxy(socket string) http.Handler {
	program := &url.URL{Scheme: "http", Host: "program"}
	proxy := httputil.NewSingleHostReverseProxy(program)
	proxy.Transport = &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		var dialer net.Dialer
		return dialer.DialContext(ctx, "unix", socket)
	}}
	return proxy
}
