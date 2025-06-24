# CacheMar - Cache Management Library

CacheMar is a versatile cache management library crafted to offer a seamless interface across multiple caching systems. With its intuitive design, developers can effortlessly transition between different caching drivers, ensuring adaptability without substantial code alterations. Presently, CacheMar extends support to both in-memory caching and Memcached, with the flexibility for future expansions.

## Table of Contents

- [CacheMar - Cache Management Library](#cachemar---cache-management-library)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Installation](#installation)
  - [Getting Started](#getting-started)
  - [Supported Drivers](#supported-drivers)
  - [Usage](#usage)
    - [Creating a CacheMar Service](#creating-a-cachemar-service)
    - [Registering Caching Drivers](#registering-caching-drivers)
    - [Setting the Current Cache Driver](#setting-the-current-cache-driver)
    - [Using the Cache](#using-the-cache)
    - [Using Tags for Invalidation](#using-tags-for-invalidation)
    - [Using Chains](#using-chains)
    - [Using Chaining](#using-chaining)
    - [Setting a Fallback](#setting-a-fallback)
    - [Overriding the Chain](#overriding-the-chain)
    - [Other Cache Operations](#other-cache-operations)
  - [Examples](#examples)
    - [In-Memory Cache Example](#in-memory-cache-example)
    - [Memcached Example](#memcached-example)
    - [Redis Example](#redis-example)
    - [Redis Cluster Example](#redis-cluster-example)
- [License](#license)
  - [Testing](#testing)
    - [Quick Test Commands](#quick-test-commands)
    - [Test Environment Setup](#test-environment-setup)
    - [Automated Testing](#automated-testing)
    - [Docker Testing Environment](#docker-testing-environment)
  - [Documentation](#documentation)
- [Contributing](#contributing)

## Features

* **Unified API**: A consistent interface across various caching drivers, making it easy to switch or combine them.
* **Multiple Drivers**: Built-in support for in-memory caching, Memcached, Redis single instance, and Redis Cluster.
* **Redis Cluster Support**: Full Redis Cluster support with automatic failover, load balancing, and advanced configuration options.
* **Dynamic Switching**: Seamlessly switch between caching drivers with minimal code changes.
* **Chaining Mechanism**: Chain multiple cache managers for a fallback mechanism. If one manager doesn't have the data or encounters an error, the next one in the chain is used.
* **Tag-based Caching**: Invalidate cache entries easily using tags.
* **Numeric Operations**: Increment and decrement operations for integer values in the cache.
* **Context Compatibility**: Fully compatible with Go's context package, allowing for request-scoped caching.
* **Fallback Support**: Set a default fallback cache manager to be used if none in the chain have the data.
* **Chain Override**: Temporarily override the chain for specific calls without affecting the original configuration.
* **Circuit Breaker Pattern**: Automatically switch to fallback cachers if the primary cacher fails, and switch back when it's available again.
* **Advanced Redis Features**: TLS/SSL support, connection pooling, compression, and flexible routing options.
* **100% Backward Compatibility**: Existing Redis single-instance code works unchanged.

## Installation
To use CacheMar in your Go project, you can install it using the go get command:

```bash
go get github.com/stremovskyy/cachemar
```

## Getting Started
Before you start using CacheMar, you need to import the package into your Go code:

```go
import (
"context"
"time"

    "github.com/stremovskyy/cachemar"
)
```

## Supported Drivers
CacheMar seamlessly integrates with a variety of caching drivers, including:

1. **In-Memory Cache**: Leveraging Go's sync.Map, this driver offers a straightforward in-memory caching solution. It's an ideal choice for applications seeking a temporary and nimble caching mechanism.
2. **Memcached**: With CacheMar, interfacing with Memcached—a renowned distributed caching system—becomes effortless. It's tailored for expansive applications necessitating cache distribution across multiple instances or servers.
3. **Redis**: CacheMar facilitates smooth interactions with Redis, a prominent in-memory data structure store. Supports both single instance and cluster modes with full backward compatibility.
   - **Single Instance**: Traditional Redis setup with one server
   - **Cluster Mode**: Distributed Redis cluster for high availability and scalability
   - **TLS Support**: Secure connections with SSL/TLS encryption
   - **Connection Pooling**: Optimized connection management for performance
   - **Compression**: Optional Gzip compression for stored data


## Usage
### Creating a CacheMar Service
To use CacheMar, you first need to create a new Service instance using cachemar.New():

```go
cacheService := cachemar.New()
```

### Registering Caching Drivers
Next, you can register the caching drivers you want to use with CacheMar:

```go
inMemoryCache := memory.NewMemoryCacheService()
cacheService.Register("in-memory", inMemoryCache)

memcachedOptions := &memcached.Options{
Servers: []string{"localhost:11211"},
Prefix:  "my-cache",
}
memcachedCache := memcached.NewCacheService(memcachedOptions)
cacheService.Register("memcached", memcachedCache)
```

### Setting the Current Cache Driver
After registering the caching drivers, you can set the current driver to use with CacheMar:

```go
cacheService.SetCurrentDriver("in-memory")
```

### Using the Cache
Once you have set the current cache driver, you can use the cache methods provided by CacheMar:

```go
ctx := context.Background()
key := "my-key"
value := "my-value"
ttl := 5 * time.Minute

// Set a value in the cache
err := cacheService.Set(ctx, key, value, ttl, nil)
if err != nil {
    // Handle error
}

// Get a value from the cache
var retrievedValue string
err = cacheService.Get(ctx, key, &retrievedValue)
if err != nil {
    // Handle error or cache miss
}

// Remove a value from the cache
err = cacheService.Remove(ctx, key)
if err != nil {
    // Handle error
}
```

### Using Tags for Invalidation
CacheMar supports tag-based caching to easily invalidate multiple keys at once:

```go
tags := []string{"tag1", "tag2"}

// Set a value in the cache with tags
err := cacheService.Set(ctx, key, value, ttl, tags)
if err != nil {
    // Handle error
}

// Remove all keys associated with a specific tag
err = cacheService.RemoveByTag(ctx, "tag1")
if err != nil {
    // Handle error
}
```

### Using Chains
CacheMar supports chaining multiple cache managers together for a fallback mechanism. If one manager doesn't have the data or encounters an error, the next one in the chain is used. You can create a chain of cache managers using cachemar.Chain():

   ```go
   manager := cachemar.New()
manager.Register("redis", redisCache)
manager.Register("memory", memoryCache)
manager.SetCurrent("redis")
   ```

### Using Chaining
With chaining, you can specify an order of cache managers. If the first one doesn't have the data, the next one is queried, and so on.
```go
chainedManager := manager.Chain()
chainedManager.AddToChain("redis")
chainedManager.AddToChain("memory")
chainedManager.Get(ctx, "somekey", &value) // This will first check redis, then memory.

```

### Setting a Fallback
A fallback cache manager can be set which will be queried if none in the chain have the data

```go
chainedManager.SetFallback("memory")
```

### Overriding the Chain
For specific calls, you can override the chain without affecting the original chain configuration.
```go
chainedManager.Override("memory", "redis").Get(ctx, "somekey", &value) // This will first check memory, then redis.
```

### Using the Circuit Breaker Pattern
CacheMar supports the circuit breaker pattern to automatically switch to fallback cachers if the primary cacher fails, and switch back when it's available again. This is useful for handling temporary failures in the primary cacher, such as network issues or server restarts.

```go
// Create a manager with circuit breaker
manager := cachemar.NewWithOptions(
    cachemar.WithDebug(),
    cachemar.WithCircuitBreaker("redis", []string{"memory"}, 5*time.Second),
)

// Register the cachers
manager.Register("redis", redisCache)
manager.Register("memory", memoryCache)

// Use the manager as usual
err := manager.Set(ctx, "key", "value", time.Minute, nil)
```

In this example, if the Redis cacher fails, the manager will automatically switch to the memory cacher. It will periodically check if the Redis cacher is back online, and switch back to it when it is.

### Other Cache Operations
CacheMar also provides other cache operations like increment and decrement for integer values:

```go
key := "my-counter"

// Increment the value of the key by 1
err := cacheService.Increment(ctx, key)
if err != nil {
    // Handle error
}

// Decrement the value of the key by 1
err = cacheService.Decrement(ctx, key)
if err != nil {
    // Handle error
}
```

## Examples
### In-Memory Cache Example

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/stremovskyy/cachemar"
    "github.com/stremovskyy/cachemar/memory"
)

func main() {
    cacheService := cachemar.New()

    inMemoryCache := memory.NewMemoryCacheService()
    cacheService.Register("in-memory", inMemoryCache)

    ctx := context.Background()
    key := "my-key"
    value := "my-value"
    ttl := 5 * time.Minute

    // Set a value in the cache
    err := cacheService.Set(ctx, key, value, ttl, nil)
    if err != nil {
        fmt.Println("Error:", err)
    }

    // Get a value from the cache
    var retrievedValue string
    err = cacheService.Get(ctx, key, &retrievedValue)
    if err != nil {
        fmt.Println("Error:", err)
    }

    fmt.Println("Retrieved Value:", retrievedValue)
}

```

### Memcached Example

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/stremovskyy/cachemar"
    "github.com/stremovskyy/cachemar/memcached"
)

func main() {
    cacheService := cachemar.New()

    memcachedOptions := &memcached.Options{
        Servers: []string{"localhost:11211"},
        Prefix:  "my-cache",
    }
    memcachedCache := memcached.NewCacheService(memcachedOptions)
    cacheService.Register("memcached", memcachedCache)

    ctx := context.Background()
    key := "my-key"
    value := "my-value"
    ttl := 5 * time.Minute

    // Set a value in the cache
    err := cacheService.Set(ctx, key, value, ttl, nil)
    if err != nil {
        fmt.Println("Error:", err)
    }

    // Get a value from the cache
    var retrievedValue string
    err = cacheService.Get(ctx, key, &retrievedValue)
    if err != nil {
        fmt.Println("Error:", err)
    }

    fmt.Println("Retrieved Value:", retrievedValue)
}

   ```
### Redis Example

```go
    package main

    import (
        "context"
        "fmt"
        "time"

        "github.com/stremovskyy/cachemar"
        "github.com/stremovskyy/cachemar/drivers/redis"
    )

    func main() {
        cacheService := cachemar.New()

        // Single Redis instance (backward compatible)
        redisOptions := redis.NewSingleInstanceOptions("localhost:6379", "", 0).
            WithCompression().
            WithPrefix("my-cache")

        redisCache := redis.New(redisOptions)
        cacheService.Register("redis", redisCache)

        ctx := context.Background()
        key := "my-key"
        value := "my-value"
        ttl := 5 * time.Minute

        // Set a value in the cache
        err := cacheService.Set(ctx, key, value, ttl, nil)
        if err != nil {
            fmt.Println("Error:", err)
        }

        // Get a value from the cache
        var retrievedValue string
        err = cacheService.Get(ctx, key, &retrievedValue)
        if err != nil {
            fmt.Println("Error:", err)
        }

        fmt.Println("Retrieved Value:", retrievedValue)
    }
```

### Redis Cluster Example

```go
    package main

    import (
        "context"
        "fmt"
        "time"

        "github.com/stremovskyy/cachemar"
        "github.com/stremovskyy/cachemar/drivers/redis"
    )

    func main() {
        cacheService := cachemar.New()

        // Redis Cluster configuration
        clusterOptions := redis.NewClusterOptions(
            []string{
                "localhost:7000",
                "localhost:7001", 
                "localhost:7002",
                "localhost:7003",
                "localhost:7004",
                "localhost:7005",
            },
            "", // cluster password
        ).WithCompression().WithPrefix("cluster-cache")

        clusterCache := redis.New(clusterOptions)
        cacheService.Register("redis-cluster", clusterCache)

        ctx := context.Background()
        key := "cluster-key"
        value := map[string]interface{}{
            "user_id": 123,
            "name":    "John Doe",
            "active":  true,
        }
        ttl := 1 * time.Hour

        // Set a value in the cluster
        err := cacheService.Set(ctx, key, value, ttl, []string{"users", "active"})
        if err != nil {
            fmt.Println("Error:", err)
        }

        // Get a value from the cluster
        var retrievedValue map[string]interface{}
        err = cacheService.Get(ctx, key, &retrievedValue)
        if err != nil {
            fmt.Println("Error:", err)
        }

        fmt.Printf("Retrieved Value: %+v\n", retrievedValue)

        // Get keys by tag
        keys, err := cacheService.GetKeysByTag(ctx, "users")
        if err != nil {
            fmt.Println("Error:", err)
        }

        fmt.Printf("User keys: %v\n", keys)
    }
```
> 📖 **For comprehensive Redis cluster documentation**, including advanced configuration options, TLS setup, performance tuning, and migration guides, see [Redis Cluster Documentation](docs/REDIS_CLUSTER.md).
        err := cacheService.Set(ctx, key, value, ttl, nil)
        if err != nil {
            fmt.Println("Error:", err)
        }

        // Get a value from the cache
        var retrievedValue string
        err = cacheService.Get(ctx, key, &retrievedValue)
        if err != nil {
            fmt.Println("Error:", err)
        }

        fmt.Println("Retrieved Value:", retrievedValue)
    }

```
### Chaining Multiple Cache Managers Example

```go
    package main

    import (
        "context"
        "fmt"
        "time"

        "github.com/stremovskyy/cachemar"
    )

    func main() {
      // Initialize CacheMar and register cache managers
      manager := cachemar.New()
      manager.Register("inMemory", inMemoryCache)
      manager.Register("memcached", memcachedCache)
      manager.Register("redis", redisCache)

      // Create a chain of cache managers
      chainedManager := manager.Chain()
      chainedManager.AddToChain("inMemory")
      chainedManager.AddToChain("memcached")
      chainedManager.AddToChain("redis")

      // Use the chained manager
      err := chainedManager.Get(ctx, "someKey", &value)

	  // Or in manager 
	  err := manager.Chain().Get(ctx, "someKey", &value)
    }

```

### Circuit Breaker Pattern Example

```go
    package main

    import (
        "context"
        "fmt"
        "time"

        "github.com/stremovskyy/cachemar"
        "github.com/stremovskyy/cachemar/drivers/memory"
        "github.com/stremovskyy/cachemar/drivers/redis"
    )

    func main() {
        // Create a manager with circuit breaker
        // This will use Redis as the primary cacher and memory as the fallback
        // It will check if Redis is back online every 5 seconds
        manager := cachemar.NewWithOptions(
            cachemar.WithDebug(),
            cachemar.WithCircuitBreaker(string(cachemar.RedisCacherName), []string{string(cachemar.MemoryCacherName)}, 5*time.Second),
        )

        // Register the Redis cacher
        redisOptions := redis.NewSingleInstanceOptions("localhost:6379", "", 0).
            WithCompression().
            WithPrefix("my-cache")
        redisCache := redis.New(redisOptions)
        manager.Register(string(cachemar.RedisCacherName), redisCache)

        // Register the memory cacher
        memoryCache := memory.New()
        manager.Register(string(cachemar.MemoryCacherName), memoryCache)

        // Use the manager as usual
        ctx := context.Background()
        key := "my-key"
        value := "my-value"
        ttl := 5 * time.Minute

        // Set a value in the cache
        // If Redis is unavailable, it will automatically use the memory cacher
        err := manager.Set(ctx, key, value, ttl, nil)
        if err != nil {
            fmt.Println("Error:", err)
        }

        // Get a value from the cache
        // If Redis is unavailable, it will automatically use the memory cacher
        var retrievedValue string
        err = manager.Get(ctx, key, &retrievedValue)
        if err != nil {
            fmt.Println("Error:", err)
        }

        fmt.Println("Retrieved Value:", retrievedValue)

        // When Redis comes back online, the manager will automatically switch back to it
    }
```

# License
CacheMar is licensed under the MIT license. See the [LICENSE](LICENSE) file for more info.

# Contributing
Contributions are welcome! Feel free to open an issue or submit a pull request if you have a way to improve CacheMar.
