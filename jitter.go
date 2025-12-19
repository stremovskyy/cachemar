package cachemar

import (
	"math/rand"
	"sync"
	"time"
)

var (
	jitterMu  sync.Mutex
	jitterRng = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func applyTTLJitter(base, maxJitter time.Duration) time.Duration {
	if base <= 0 || maxJitter <= 0 {
		return base
	}
	if maxJitter > base {
		maxJitter = base
	}

	jitterMu.Lock()
	n := jitterRng.Int63n(int64(maxJitter) + 1)
	jitterMu.Unlock()

	return base + time.Duration(n)
}
