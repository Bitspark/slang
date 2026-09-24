package capability

import (
	"fmt"
	"net"
	"syscall"
)

// PrivateNetworks are the address ranges a program must not reach through a
// capability host: unspecified, loopback, private, shared, link-local (which
// includes cloud metadata services), documentation, benchmarking, multicast and
// reserved ranges, for IPv4 and IPv6.
var PrivateNetworks = mustParseNetworks(
	"0.0.0.0/8", "10.0.0.0/8", "100.64.0.0/10", "127.0.0.0/8", "169.254.0.0/16",
	"172.16.0.0/12", "192.0.0.0/24", "192.0.2.0/24", "192.168.0.0/16", "198.18.0.0/15",
	"198.51.100.0/24", "203.0.113.0/24", "224.0.0.0/4", "240.0.0.0/4",
	"::/128", "::1/128", "64:ff9b::/96", "100::/64", "2001:db8::/32", "fc00::/7",
	"fe80::/10", "ff00::/8",
)

func mustParseNetworks(cidrs ...string) []*net.IPNet {
	networks, err := ParseNetworks(cidrs)
	if err != nil {
		panic(err)
	}
	return networks
}

// ParseNetworks parses CIDR notation, such as "192.0.2.1/32".
func ParseNetworks(cidrs []string) ([]*net.IPNet, error) {
	networks := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}
		networks = append(networks, network)
	}
	return networks, nil
}

// RefuseAddresses returns a net.Dialer Control function that refuses connections
// to any address in refused. It runs on the address actually dialed, after name
// resolution, so a name that resolves to a refused address is refused as well, and
// every redirect is checked again when it dials. An IPv4 address written as IPv6
// is checked as IPv4.
func RefuseAddresses(refused []*net.IPNet) func(network, address string, _ syscall.RawConn) error {
	return func(network, address string, _ syscall.RawConn) error {
		host, _, err := net.SplitHostPort(address)
		if err != nil {
			return err
		}
		ip := net.ParseIP(host)
		if ip == nil {
			return fmt.Errorf("refused %s: not an IP address", address)
		}
		if v4 := ip.To4(); v4 != nil {
			ip = v4
		}
		for _, network := range refused {
			if network.Contains(ip) {
				return fmt.Errorf("refused %s: %s is not reachable from programs", address, network)
			}
		}
		return nil
	}
}
