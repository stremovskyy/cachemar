# Cachemar Service

## Description
Cachemar Service is a versatile caching service in Go that abstracts away the underlying cache mechanism, offering a simplified and consistent interface for caching needs. Currently, the service includes a driver for Redis and has plans to include in-memory and Memcached caching strategies.

## Features

1. Easy to Use: Cachemar provides a simple yet powerful interface for dealing with caching in your applications.
2. Redis Driver: The service supports Redis, a robust, open-source, in-memory data structure store used as a database, cache, and message broker.
3. Goroutine Safe: The service is built with concurrency in mind and is safe for use with goroutines.
4. Gzip Compression: The Redis driver uses Gzip compression for storing data, providing efficient memory usage especially for larger data sets.
5. TTL Support: The service supports setting a TTL (time to live) for cached data, allowing for automatic expiration of data.
6. Key Prefixing: The service supports key prefixing, allowing for easy grouping of cached data.
7. Key Generation: The service supports key generation, allowing for easy creation of unique keys.

## Installation

```bash
go get github.com/alextanhongpin/cachemar
```

## Usage

```go

options := &cachemar.RedisCacheServiceOptions{
	DSN:                "localhost:6379",
	Password:           "", // Set password if required
	Database:           0,
	CompressionEnabled: true,
	Prefix:             "myapp",
}

cacheService := cachemar.NewCacheService(options)
cacheService.Set("key", "value", 0)
value, err := cacheService.Get("key")
```

## Roadmap
1. In-Memory Caching: As an important feature for situations where high-speed access to cached data is crucial, in-memory caching will be added soon.
2. Memcached Driver: Memcached is a general-purpose distributed memory-caching system that we plan to support as an additional driver.
3. Additional Drivers: We plan to add additional drivers for other caching mechanisms, such as BoltDB, LevelDB, and more.
4. Additional Features: We plan to add additional features, such as support for multiple cache instances, cache tagging, and more.

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License
Cachemar Service is licensed under the [MIT](https://choosealicense.com/licenses/mit/) license.
