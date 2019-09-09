# ratelimit

A distributed ratelimiter based on lua + redis

## Usage


``` 
package main

import (
	"github.com/seaguest/ratelimit"
)

func main() {
	ratelimit.Init("127.0.0.1:6379", "")

	ratelimit.Take("1", 60, 100000, 1)
}

```
