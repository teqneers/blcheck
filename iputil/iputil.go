package iputil

import (
	"net"
	s "strings"
)

// Validate checks if given IP is a valid one
func Validate(toCheckIP string) bool {
	return net.ParseIP(toCheckIP) != nil
}

// Reverse flips an IP to put into blacklist request
func Reverse(ip string) string {
	ipParts := s.Split(ip, ".")

	for i, j := 0, len(ipParts)-1; i < j; i, j = i+1, j-1 {
		ipParts[i], ipParts[j] = ipParts[j], ipParts[i]
	}

	return s.Join(ipParts, ".")
}
