package cachemar

type Option func(*manager)

func WithDebug() Option {
	return func(m *manager) {
		m.debug = true
	}
}
