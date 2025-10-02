package tests_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stremovskyy/cachemar"
	"github.com/stremovskyy/cachemar/drivers/memory"
)

type memoryStressConfig struct {
	name        string
	maxSize     int
	workingSet  int
	valueBytes  int
	valueSizes  []int
	ttl         time.Duration
	writeEvery  int
	tagEvery    int
	parallelism int
}

func BenchmarkMemoryLRUStress(b *testing.B) {
	configs := []memoryStressConfig{
		{
			name:        "HotSet-Unbounded",
			maxSize:     0,
			workingSet:  1 << 16,
			valueBytes:  128,
			ttl:         0,
			writeEvery:  32,
			parallelism: 4,
		},
		{
			name:        "Eviction-Pressure",
			maxSize:     1 << 13,
			workingSet:  1 << 15,
			valueBytes:  512,
			ttl:         time.Minute,
			writeEvery:  5,
			tagEvery:    16,
			parallelism: 8,
		},
		{
			name:        "Short-TTL-Churn",
			maxSize:     1 << 12,
			workingSet:  1 << 14,
			valueBytes:  256,
			ttl:         250 * time.Millisecond,
			writeEvery:  3,
			tagEvery:    8,
			parallelism: 8,
		},
		{
			name:        "Million-Varied-Sizes",
			maxSize:     1 << 18,
			workingSet:  (1 << 20) + (1 << 18),
			valueSizes:  []int{64, 128, 256, 512, 1024},
			ttl:         5 * time.Second,
			writeEvery:  2,
			tagEvery:    32,
			parallelism: 16,
		},
		{
			name:        "TenMillion-Mixed-Heavy",
			maxSize:     1 << 11,
			workingSet:  10_500_000,
			valueSizes:  []int{1 * 1024, 4 * 1024, 16 * 1024, 48 * 1024, 80 * 1024, 100 * 1024},
			ttl:         2 * time.Second,
			writeEvery:  1,
			tagEvery:    64,
			parallelism: 32,
		},
		{
			name:        "Millions-Mixed-Heavy",
			maxSize:     1 << 11,
			workingSet:  100_500_000,
			valueSizes:  []int{1 * 1024, 4 * 1024, 16 * 1024, 48 * 1024, 80 * 1024, 100 * 1024, 10 * 1024, 40 * 1024, 160 * 1024, 480 * 1024, 800 * 1024, 1000 * 1024},
			ttl:         2 * time.Second,
			writeEvery:  1,
			tagEvery:    64,
			parallelism: 32,
		},
	}

	for _, cfg := range configs {
		cfg := cfg
		b.Run(
			cfg.name, func(b *testing.B) {
				runMemoryLRUStressBenchmark(b, cfg)
			},
		)
	}
}

func runMemoryLRUStressBenchmark(b *testing.B, cfg memoryStressConfig) {
	if cfg.workingSet <= 0 {
		b.Fatal("workingSet must be positive")
	}
	if cfg.valueBytes <= 0 && len(cfg.valueSizes) == 0 {
		b.Fatal("either valueBytes or valueSizes must be provided")
	}

	cache := memory.NewWithConfig(memory.Config{MaxSize: cfg.maxSize})
	ctx := context.Background()
	totalKeys := cfg.workingSet
	makeKey := func(i int) string {
		return fmt.Sprintf("stress-key-%d", i)
	}

	templates := make(map[int][]byte)
	buildTemplate := func(size int) []byte {
		if buf, ok := templates[size]; ok {
			return buf
		}
		buf := make([]byte, size)
		for i := range buf {
			buf[i] = byte(i % 251)
		}
		templates[size] = buf
		return buf
	}

	chooseValueSize := func(pos int) int {
		if len(cfg.valueSizes) == 0 {
			return cfg.valueBytes
		}
		return cfg.valueSizes[pos%len(cfg.valueSizes)]
	}

	prefillCount := cfg.workingSet
	if cfg.maxSize > 0 && cfg.maxSize < prefillCount {
		prefillCount = cfg.maxSize
	}

	for i := 0; i < prefillCount; i++ {
		size := chooseValueSize(i)
		if err := cache.Set(ctx, makeKey(i), buildTemplate(size), cfg.ttl, nil); err != nil {
			b.Fatalf("prefill set failed: %v", err)
		}
	}

	var opSequence uint64
	var benchErr atomic.Pointer[error]

	if cfg.parallelism > 0 {
		b.SetParallelism(cfg.parallelism)
	}

	bytesPerOp := cfg.valueBytes
	if len(cfg.valueSizes) > 0 {
		for _, size := range cfg.valueSizes {
			if size > bytesPerOp {
				bytesPerOp = size
			}
		}
	}
	if bytesPerOp > 0 {
		b.SetBytes(int64(bytesPerOp))
	}
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(
		func(pb *testing.PB) {
			var localValue []byte
			for pb.Next() {
				if benchErr.Load() != nil {
					return
				}

				sequence := atomic.AddUint64(&opSequence, 1)
				idx := int(sequence % uint64(totalKeys))
				key := makeKey(idx)
				withTags := cfg.tagEvery > 0 && sequence%uint64(cfg.tagEvery) == 0

				if cfg.writeEvery > 0 && sequence%uint64(cfg.writeEvery) == 0 {
					// Reuse a private slice to avoid sharing between goroutines
					size := chooseValueSize(int(sequence))
					template := buildTemplate(size)
					if cap(localValue) < size {
						localValue = make([]byte, size)
					}
					localValue = localValue[:size]
					copy(localValue, template)

					var tags []string
					if withTags {
						tags = []string{fmt.Sprintf("tag-%d", idx%64)}
					}

					if err := cache.Set(ctx, key, localValue, cfg.ttl, tags); err != nil {
						errCopy := err
						benchErr.CompareAndSwap(nil, &errCopy)
						return
					}
				} else {
					var payload []byte
					err := cache.Get(ctx, key, &payload)
					if err != nil && err != cachemar.ErrNotFound {
						errCopy := err
						benchErr.CompareAndSwap(nil, &errCopy)
						return
					}

					// Periodically exercise tag eviction paths during reads
					if withTags && err == nil {
						err = cache.RemoveByTag(ctx, fmt.Sprintf("tag-%d", idx%64))
						if err != nil {
							errCopy := err
							benchErr.CompareAndSwap(nil, &errCopy)
							return
						}
					}
				}
			}
		},
	)

	b.StopTimer()

	if errPtr := benchErr.Load(); errPtr != nil {
		b.Fatalf("stress harness error: %v", *errPtr)
	}
}
