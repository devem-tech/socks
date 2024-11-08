package main

import (
	"time"

	"github.com/devem-tech/statsd"

	"github.com/devem-tech/socks/internal/server"
	"github.com/devem-tech/socks/pkg/resolver"
)

const dnsCacheTTL = 10 * time.Minute

func main() {
	// Create a StatsD client
	metrics, err := statsd.New()
	if err != nil {
		panic(err)
	}
	defer metrics.Close()

	// Create a DNS resolver
	dnsResolver := resolver.New(metrics, dnsCacheTTL)

	// Create a proxy server
	s := server.New(
		server.Network("tcp"),
		server.Address(":7010"),
		server.DNSResolver(dnsResolver),
		server.Metrics(metrics),
	)

	// Run the server
	s.Serve()
}
