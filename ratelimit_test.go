package ratelimit_test

import (
	"github.com/seaguest/log"
	"github.com/seaguest/ratelimit"
	"testing"
	"time"
)

/*
func BenchmarkFoo(b *testing.B) {
	Init("127.0.0.1:6379", "")

	for i := 0; i < b.N; i++ {
		Allow("1", 60, 100)
	}
}
*/
func Test_A(b *testing.T) {

	log.Error("xxx")
	ratelimit.Init("172.17.0.5:6379", "")

	for i := 0; i < 10; i++ {
		c, err := ratelimit.Allow("1", 10, 10, 1)
		log.Error(c, err)
	}

	for i := 0; i < 20; i++ {
		c, err := ratelimit.Allow("1", 10, 10, 1)
		time.Sleep(time.Millisecond * 500)
		log.Error(c, err)
	}
}
