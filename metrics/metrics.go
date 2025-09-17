package metrics

import "time"

var stats map[string]time.Duration

func init() {
	stats = make(map[string]time.Duration, 32)
}

func Track(name string, fn func()) {
	start := time.Now()
	fn()
	stats[name] = time.Since(start)
}

func Get() map[string]time.Duration {
	return stats
}

