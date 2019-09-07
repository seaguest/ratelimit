package ratelimit

import (
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
)

func GetRedisPool(address, password string, maxConnection int) *redis.Pool {
	pool := &redis.Pool{
		MaxIdle:     maxConnection,
		MaxActive:   maxConnection,
		Wait:        false,
		IdleTimeout: 240 * time.Second,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		Dial: func() (redis.Conn, error) {
			return dial("tcp", address, password)
		},
	}
	return pool
}

func dial(network, address, password string) (redis.Conn, error) {
	c, err := redis.Dial(network, address)
	if err != nil {
		return nil, err
	}
	if password != "" {
		if _, err := c.Do("AUTH", password); err != nil {
			c.Close()
			return nil, err
		}
	}
	return c, err
}

func BenchmarkFoo(b *testing.B) {
	pool := GetRedisPool("127.0.0.1:6379", "", 10000)

	for i := 0; i < b.N; i++ {
		Take("1", 60, 100000, 1, pool)
	}
}
