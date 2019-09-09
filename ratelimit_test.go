package ratelimit

import (
	"testing"
)

func BenchmarkFoo(b *testing.B) {
	Init("127.0.0.1:6379", "")

	for i := 0; i < b.N; i++ {
		Take("1", 60, 100000, 1)
	}
}
