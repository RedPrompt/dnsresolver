package runner

import "flag"

type Options struct {
	Domain      string
	DomainsList string
	Threads     int
	Timeout     int
	ShowIPs     bool
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
