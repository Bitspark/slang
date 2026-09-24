package capability

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRefuseAddressesCoversPrivateAndReservedRanges(t *testing.T) {
	refuse := RefuseAddresses(PrivateNetworks)
	for address, refused := range map[string]bool{
		"127.0.0.1:80":            true,
		"10.1.2.3:443":            true,
		"172.30.0.1:80":           true, // the programs' own Docker network
		"192.168.1.1:80":          true,
		"169.254.169.254:80":      true, // cloud metadata
		"100.64.0.1:80":           true,
		"0.0.0.0:80":              true,
		"224.0.0.1:80":            true,
		"[::1]:80":                true,
		"[fe80::1]:80":            true,
		"[fd00::1]:80":            true,
		"[::ffff:10.0.0.1]:80":    true, // IPv4-mapped IPv6 is checked as IPv4
		"[64:ff9b::a00:1]:80":     true, // NAT64 can reach IPv4 addresses
		"93.184.215.14:443":       false,
		"1.1.1.1:53":              false,
		"[2606:4700:4700::1]:443": false,
	} {
		err := refuse("tcp", address, nil)
		if refused {
			require.Error(t, err, address)
		} else {
			require.NoError(t, err, address)
		}
	}
}

func TestRefuseAddressesAcceptsAdditionalNetworks(t *testing.T) {
	extra, err := ParseNetworks([]string{"198.51.100.7/32"})
	require.NoError(t, err)
	refuse := RefuseAddresses(extra)
	require.Error(t, refuse("tcp", "198.51.100.7:443", nil))
	require.NoError(t, refuse("tcp", "198.51.100.8:443", nil))
	_, err = ParseNetworks([]string{"not-a-network"})
	require.Error(t, err)
}

// A host name is resolved before the check, so it cannot smuggle a refused address.
func TestHTTPHostRefusesANameThatResolvesToARefusedAddress(t *testing.T) {
	var reached int64
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&reached, 1)
		io.WriteString(w, "reached")
	}))
	defer target.Close()
	_, port, err := net.SplitHostPort(strings.TrimPrefix(target.URL, "http://"))
	require.NoError(t, err)

	dialer := &net.Dialer{Control: RefuseAddresses(PrivateNetworks)}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = dialer.DialContext
	client, _ := startHost(t, &HTTPHost{Client: &http.Client{Transport: transport}})

	_, err = client.Get("http://localhost:" + port)
	require.ErrorContains(t, err, "502")
	require.ErrorContains(t, err, "not reachable from programs")
	require.Zero(t, atomic.LoadInt64(&reached))
}
