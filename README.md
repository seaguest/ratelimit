# ratelimit

A distributed ratelimiter based on redis

## Usage


``` 
package main

import (
	"github.com/seaguest/ratelimit"
)

func main() {
		ratelimit.Take("key", 1, 10, 1, pool)
}

```
