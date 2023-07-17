package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

type Options struct {
	Domain      string // Domain name to validate DNS
	DomainsList string // File containing list of domains to validate DNS
	// Threads     int
}

func checkDNS(domain string, wg *sync.WaitGroup) {

	// Get each domain from appended domain slice
	// Resolve DNS for each domain and print if resolved
	_, err := net.ResolveIPAddr("ip4", domain)
	// fmt.Println("Checking: ", domain)
	if err == nil {
		fmt.Println(domain)
	}

	wg.Done()
}

func main() {
	// Set options
	options := &Options{}

	flag.StringVar(&options.Domain, "d", "", "Domain name to validate DNS")
	flag.StringVar(&options.DomainsList, "dl", "", "File containing list of domains to validate DNS")
	// flag.IntVar(&options.Threads, "t", 10, "Number of concurrent gorountines to resolve DNS")
	flag.Parse()

	// Declare domains slice
	var domainSlice []string

	// Check and get Stdin
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Get input from stdin
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			domainSlice = append(domainSlice, sc.Text())
		}
	}

	// Get domain from Domain flag
	if options.Domain != "" {
		domainSlice = append(domainSlice, options.Domain)
	}

	// Get domains from DomainList flag
	if options.DomainsList != "" {
		readFile, err := os.Open(options.DomainsList)
		if err != nil {
			fmt.Println(err)
		}

		fileScanner := bufio.NewScanner(readFile)
		fileScanner.Split(bufio.ScanLines)

		for fileScanner.Scan() {
			domainSlice = append(domainSlice, fileScanner.Text())
		}

		readFile.Close()
	}

	var wg sync.WaitGroup
	for _, domain := range domainSlice {
		wg.Add(1)
		go checkDNS(domain, &wg)
		// 10 Goroutines per second
		time.Sleep(100 * time.Millisecond)
	}
	wg.Wait()

}
