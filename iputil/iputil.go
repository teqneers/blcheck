package iputil

import (
	"net"
	"regexp"
	"strings"
)

var (
	// RegexIP to check for ips
	RegexIP = regexp.MustCompile(`([0-9]{1,3}).([0-9]{1,3}).([0-9]{1,3}).([0-9]{1,3})`)
	// RegexDomain to check for domains
	RegexDomain = regexp.MustCompile(`([a-zA-Z0-9]+(-[a-zA-Z0-9]+)*.)+[a-zA-Z]{2,}`)
	//RegexTdl to check for tdl
	RegexTdl = regexp.MustCompile(`([a-zA-Z0-9]+(-[a-zA-Z0-9]+)*.)[a-zA-Z]{2,}$`)
)

// ValidateIP checks if given IP is a valid one
func ValidateIP(toCheckIP string) bool {
	return net.ParseIP(toCheckIP) != nil && RegexIP.MatchString(toCheckIP)
}

// IsIP determines if given string is an IP
func IsIP(toCheckIP string) bool {
	return RegexIP.MatchString(toCheckIP)
}

// IsHostname determines if given string is a hostname
func IsHostname(toCheckHostname string) bool {
	return RegexDomain.MatchString(toCheckHostname)
}

// Reverse flips an IP or hostname to put into blacklist request
func Reverse(ip string) string {
	parts := strings.Split(ip, ".")

	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	return strings.Join(parts, ".")
}
