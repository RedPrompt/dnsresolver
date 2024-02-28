package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

type ResolvedHost struct {
	Hostname string
	IPv4     []net.IP
	IPv6     []net.IP
}

type Options struct {
	Domain      string
	DomainsList string
	Threads     int
	Timeout     int
	ShowIPs     bool
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

func parseOptions() *Options {
	options := &Options{}

	flag.StringVar(&options.Domain, "d", "", "Domain name to validate DNS")
	flag.StringVar(&options.DomainsList, "dl", "", "File containing a list of domains to validate DNS")
	flag.IntVar(&options.Threads, "t", 10, "Number of concurrent goroutines to resolve DNS")
	flag.IntVar(&options.Timeout, "timeout", 60, "Set timeout for DNS resolution in seconds")
	flag.BoolVar(&options.ShowIPs, "show-ips", false, "Show IPv4 and IPv6 addresses of the host domain")
	flag.Parse()

	return options
}

func main() {
	options := parseOptions()

	var domainSlice []string
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			domainSlice = append(domainSlice, sc.Text())
		}
	}

	if options.Domain != "" {
		domainSlice = append(domainSlice, options.Domain)
	}

	if options.DomainsList != "" {
		readFile, err := os.Open(options.DomainsList)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer readFile.Close()

		fileScanner := bufio.NewScanner(readFile)
		fileScanner.Split(bufio.ScanLines)

		for fileScanner.Scan() {
			domainSlice = append(domainSlice, fileScanner.Text())
		}
	}

	resultChan := make(chan ResolvedHost, len(domainSlice))
	concurrency := make(chan struct{}, options.Threads)
	var wg sync.WaitGroup

	for _, domain := range domainSlice {
		wg.Add(1)
		concurrency <- struct{}{}
		go func(d string) {
			defer func() { <-concurrency }()
			checkDNS(d, options.Timeout, &wg, resultChan)
		}(domain)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		fmt.Printf("%s\n", result.Hostname)

		if options.ShowIPs {
			if len(result.IPv4) > 0 {
				fmt.Println("IPv4 addresses:")
				for _, ip := range result.IPv4 {
					fmt.Printf("\t- %s\n", ip)
				}
			}

			if len(result.IPv6) > 0 {
				fmt.Println("IPv6 addresses:")
				for _, ip := range result.IPv6 {
					fmt.Printf("\t- %s\n", ip)
				}
			}
			fmt.Println()
		}
	}
}
