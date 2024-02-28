package runner

import (
	"context"
	"net"
	"sync"
	"time"
)

type ResolvedHost struct {
	Hostname string
	IPv4     []net.IP
	IPv6     []net.IP
}

var ipAddrPool = &sync.Pool{
	New: func() interface{} {
		return new(net.IPAddr)
	},
}

func checkDNS(domain string, timeout int, wg *sync.WaitGroup, resultChan chan<- ResolvedHost) {
	defer wg.Done()

	ipAddr := ipAddrPool.Get().(*net.IPAddr)
	defer ipAddrPool.Put(ipAddr)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	addrs, err := net.DefaultResolver.LookupHost(ctx, domain)
	if err != nil {
		return
	}

	var ipv4, ipv6 []net.IP
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip == nil {
			continue
		}
		if ip.To4() != nil {
			ipv4 = append(ipv4, ip)
		} else {
			ipv6 = append(ipv6, ip)
		}
	}

	resolvedHost := ResolvedHost{
		Hostname: domain,
		IPv4:     ipv4,
		IPv6:     ipv6,
	}

	resultChan <- resolvedHost
}
