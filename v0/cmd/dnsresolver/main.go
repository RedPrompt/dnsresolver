package main

import (
	"github.com/redprompt/dnsresolver/v0/pkg/runner"
	"bufio"
	"fmt"
	"os"
	"sync"
)

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
