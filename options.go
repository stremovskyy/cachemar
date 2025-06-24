package cachemar

import "time"

type Option func(*manager)

func WithDebug() Option {
	return func(m *manager) {
		m.debug = true
	}
}

// WithCircuitBreaker enables the circuit breaker pattern with the specified primary cacher,
// fallback cachers, and check interval. If the primary cacher fails, the manager will
// automatically switch to the first available fallback cacher. It will periodically check
// if the primary cacher is back online and switch back to it when it is.
//
// primaryCacher: The name of the primary cache manager
// fallbackCachers: The names of the fallback cache managers in order of preference
// checkInterval: How often to check if the primary cacher is back online
func WithCircuitBreaker(primaryCacher string, fallbackCachers []string, checkInterval time.Duration) Option {
	return func(m *manager) {
		m.useCircuitBreaker = true
		m.primaryCacher = primaryCacher
		m.fallbackCachers = fallbackCachers
		m.checkInterval = checkInterval
		m.lastCheckTime = time.Now()

		m.current = primaryCacher
	}
}
