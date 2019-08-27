# ratelimit
A ratelimit based on redis

TODO: Integrate mem + redis to reduce redis's load.

Usage:


``` 
package main

import (
	"github.com/seaguest/ratelimit"
)

func main() {
		ratelimit.Take("key", 1, 10, 1, pool)
}

```
