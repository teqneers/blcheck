package dnsutil

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	l "github.com/teqneers/blcheck/logutil"
)

// LookupIPFromHostname gets the ip for the provided hostname
func LookupIPFromHostname(lookupDomain string, dnsTimeout int) string {
	var resultIP string

	resultChan := make(chan bool, 1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dnsTimeout+1)*time.Second)

	defer cancel()

	ips, _ := net.DefaultResolver.LookupIPAddr(ctx, lookupDomain)
	resultChan <- true

	select {
	case <-resultChan:
		if len(ips) == 1 {
			resultIP = ips[0].IP.String()
		} else if len(ips) > 1 {
			var ipList []string

			for _, ip := range ips {
				ipList = append(ipList, ip.String())
			}

			l.LogError(0, fmt.Sprintf("The reverse lookup for the host %s returned multiple IPs: %s", lookupDomain, strings.Join(ipList, ", ")))
			os.Exit(155)
		} else {
			l.LogError(0, fmt.Sprintf("The reverse lookup for the host %s failed.", lookupDomain))
			os.Exit(155)
		}

	case <-ctx.Done():
		l.LogError(0, fmt.Sprintf("The reverse lookup for the host %s timed out after %d seconds.", lookupDomain, time.Duration(dnsTimeout+1)*time.Second))
		os.Exit(155)
	}

	return resultIP
}

// LookupPtrRecordForIP gets the PTR Record for the provided iP
func LookupPtrRecordForIP(lookupIP string, dnsTimeout int) string {
	var resultPTR string

	resultChan := make(chan bool, 1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dnsTimeout+1)*time.Second)

	defer cancel()

	ptrs, err := net.DefaultResolver.LookupAddr(ctx, lookupIP)
	resultChan <- true

	select {
	case <-resultChan:
		if len(ptrs) == 1 {
			resultPTR = ptrs[0]

		} else if len(ptrs) > 1 {
			l.LogError(0, fmt.Sprintf("The PTR lookup for the ip %s returned multiple PTRs: %s", lookupIP, strings.Join(ptrs, ",")))
			os.Exit(155)
		} else {
			l.LogError(0, fmt.Sprintf("The PTR lookup for the ip %s failed with: %s", lookupIP, err))
			os.Exit(155)
		}

	case <-ctx.Done():
		l.LogError(0, fmt.Sprintf("The PTR lookup for the ip %s timed out after %d seconds.", lookupIP, time.Duration(dnsTimeout+1)*time.Second))
		os.Exit(155)
	}

	return resultPTR
}
