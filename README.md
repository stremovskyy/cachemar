# CacheMar - Cache Management Library

CacheMar is a cache management library designed to provide a unified and easy-to-use interface for working with various caching systems. It allows developers to switch between different caching drivers without changing their code significantly. The library currently supports two caching drivers: an in-memory cache and Memcached.


## Table of Contents

* Features 
* Installation
* Getting Started
* Supported Drivers
* Usage
* Examples
* Contributing
* License

## Features
* Unified and consistent API for different caching drivers.
* Support for in-memory caching and Memcached.
* Ability to easily switch between caching drivers with minimal code changes.
* Tag-based caching for easy invalidation.
* Increment and decrement operations for integer values.
* Compatibility with Go's context package for request-scoped caching.

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
Currently, CacheMar supports the following caching drivers:

1. In-Memory Cache: A simple in-memory cache implemented using Go's sync.Map. This driver is suitable for applications where a temporary and lightweight caching solution is needed.
2. Memcached: CacheMar provides an interface to interact with Memcached, a distributed caching system. This driver is suitable for larger-scale applications that require caching across multiple instances or machines.
3. Redis: CacheMar provides an interface to interact with Redis, an in-memory data structure store. This driver is suitable for larger-scale applications that require caching across multiple instances or machines.

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
        "github.com/stremovskyy/cachemar/redis"
    )

    func main() {
        cacheService := cachemar.New()

        redisOptions := &redis.Options{
            Addr:     "localhost:6379",
            Password: "",
            DB:       0,
            Prefix:   "my-cache",
        }
        redisCache := redis.NewCacheService(redisOptions)
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

# License
CacheMar is licensed under the MIT license. See the [LICENSE](LICENSE) file for more info.

# Contributing
Contributions are welcome! Feel free to open an issue or submit a pull request if you have a way to improve CacheMar.
