package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
)

type Options struct {
	Domain				string				// Domain name to validate DNS
	DomainsList			string				// File containing list of domains to validate DNS
}

func main() {
	options := &Options{}

	flag.StringVar(&options.Domain, "d", "", "Domain name to validate DNS")
	flag.StringVar(&options.DomainsList, "dl", "", "File containing list of domains to validate DNS")
	flag.Parse()

	// Declare domains slice
	var domainSlice []string

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

	// Get each domain from appended domain slice
	// Resolve DNS for each domain in slice and print if resolved
	for _, domain := range domainSlice {
		_, err := net.ResolveIPAddr("ip4", domain)
		if err == nil {
			fmt.Println(domain)
		}
	}
	
}
