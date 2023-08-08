# CacheMar - Cache Management Library

CacheMar is a versatile cache management library crafted to offer a seamless interface across multiple caching systems. With its intuitive design, developers can effortlessly transition between different caching drivers, ensuring adaptability without substantial code alterations. Presently, CacheMar extends support to both in-memory caching and Memcached, with the flexibility for future expansions.

## Table of Contents

* [Features](#features)
* [Installation](#installation)
* [Getting Started](#getting-started)
* [Supported Drivers](#supported-drivers)
* [Usage](#usage)
    * [Creating a CacheMar Service](#creating-a-cachemar-service)
    * [Registering Caching Drivers](#registering-caching-drivers)
    * [Setting the Current Cache Driver](#setting-the-current-cache-driver)
    * [Using the Cache](#using-the-cache)
    * [Using Tags for Invalidation](#using-tags-for-invalidation)
    * [Using Chains](#using-chains)
    * [Setting a Fallback](#setting-a-fallback)
    * [Overriding the Chain](#overriding-the-chain)
    * [Other Cache Operations](#other-cache-operations)
* [Examples](#examples)
    * [In-Memory Cache Example](#in-memory-cache-example)
    * [Memcached Example](#memcached-example)
    * [Redis Example](#redis-example)
    * [Using Tags for Invalidation](#using-tags-for-invalidation)
* [Contributing](#contributing)
* [License](#license)

## Features

* **Unified API**: A consistent interface across various caching drivers, making it easy to switch or combine them.
* **Multiple Drivers**: Built-in support for in-memory caching, Memcached, and more. Extendable for other caching solutions.
* **Dynamic Switching**: Seamlessly switch between caching drivers with minimal code changes.
* **Chaining Mechanism**: Chain multiple cache managers for a fallback mechanism. If one manager doesn't have the data or encounters an error, the next one in the chain is used.
* **Tag-based Caching**: Invalidate cache entries easily using tags.
* **Numeric Operations**: Increment and decrement operations for integer values in the cache.
* **Context Compatibility**: Fully compatible with Go's context package, allowing for request-scoped caching.
* **Fallback Support**: Set a default fallback cache manager to be used if none in the chain have the data.
* **Chain Override**: Temporarily override the chain for specific calls without affecting the original configuration.


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
3. **Redis**: CacheMar also facilitates smooth interactions with Redis, a prominent in-memory data structure store. Like Memcached, it's apt for large-scale applications aiming for distributed caching solutions.


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

# License
CacheMar is licensed under the MIT license. See the [LICENSE](LICENSE) file for more info.

# Contributing
Contributions are welcome! Feel free to open an issue or submit a pull request if you have a way to improve CacheMar.
