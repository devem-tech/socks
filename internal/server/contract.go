package server

import (
	"fmt"
	"net"

	"github.com/devem-tech/statsd"
)

type dnsResolver interface {
	Resolve(host string) (net.IP, error)
}

type defaultDNSResolver struct{}

func (r *defaultDNSResolver) Resolve(host string) (net.IP, error) {
	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return nil, fmt.Errorf("resolve host %s: %w", host, err)
	}

	return ips[0], nil
}

type metrics interface {
	Count(key string, value int64, tags ...statsd.Tag)
	Increment(key string, tags ...statsd.Tag)
	Gauge(key string, value float64, tags ...statsd.Tag)
	Timer(key string, tags ...statsd.Tag) func()
}

type defaultMetrics struct{}

func (m *defaultMetrics) Count(_ string, _ int64, _ ...statsd.Tag) {}

func (m *defaultMetrics) Increment(_ string, _ ...statsd.Tag) {}

func (m *defaultMetrics) Gauge(_ string, _ float64, _ ...statsd.Tag) {}

func (m *defaultMetrics) Timer(_ string, _ ...statsd.Tag) func() {
	return func() {}
}
