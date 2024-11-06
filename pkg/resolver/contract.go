package resolver

import "github.com/devem-tech/statsd"

type metrics interface {
	Count(key string, value int64, tags ...statsd.Tag)
	Increment(key string, tags ...statsd.Tag)
	Gauge(key string, value float64, tags ...statsd.Tag)
	Timer(key string, tags ...statsd.Tag) func()
}
