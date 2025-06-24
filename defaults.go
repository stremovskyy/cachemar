package cachemar

import "time"

const (
	DefaultCacheTime     = time.Hour
	DefaultCheckInterval = time.Second * 5
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
