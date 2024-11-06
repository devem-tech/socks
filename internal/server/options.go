package server

type Option func(*options)

func Network(network string) Option {
	return func(o *options) {
		o.network = network
	}
}

func Address(address string) Option {
	return func(o *options) {
		o.address = address
	}
}

func DNSResolver(dnsResolver dnsResolver) Option {
	return func(o *options) {
		o.dnsResolver = dnsResolver
	}
}

func Metrics(metrics metrics) Option {
	return func(o *options) {
		o.metrics = metrics
	}
}
