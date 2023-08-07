package cachemar

import "time"

const (
	DefaultCacheTime = time.Hour
)

type CacherName string

const (
	MemoryCacherName    CacherName = "memory"
	RedisCacherName     CacherName = "redis"
	MemcachedCacherName CacherName = "memcached"
)

func (c CacherName) String() string {
	return string(c)
}
