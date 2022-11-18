package ratelimit

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

const script = `
-- key
local key = KEYS[1]
-- duration, in second
local duration = tonumber(ARGV[1])
-- max number of tokens per duration
local limit = tonumber(ARGV[2])
-- required tokens
local n = tonumber(ARGV[3])

local micro_seconds_in_second = 1000000

-- current time
local time = redis.call('time')
local now_micros = tonumber(time[1]) * micro_seconds_in_second + tonumber(time[2])

local next_free_ticket_micros = tonumber(redis.call('hget', key, 'next_free_ticket_micros') or now_micros)

-- current stored tokens
local stored_permits = tonumber(redis.call('hget', key, 'stored_permits') or limit)
-- fill interval per token
local stable_interval_micros = micro_seconds_in_second * duration / limit

-- produce token
if (now_micros > next_free_ticket_micros) then
    local new_permits = (now_micros - next_free_ticket_micros) / stable_interval_micros
    stored_permits = math.min(limit, stored_permits + new_permits)
    next_free_ticket_micros = now_micros
end

local allowed = 0
if stored_permits >= n then 
	allowed = 1
	stored_permits = stored_permits -n
end

-- consume token
redis.replicate_commands()
redis.call('hset', key, 'next_free_ticket_micros', next_free_ticket_micros)
redis.call('hset', key, 'stored_permits', stored_permits)
redis.call('expire', key, duration)

return allowed
`

var (
	pool *redis.Pool
)

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

func Init(address, password string) {
	pool = &redis.Pool{
		MaxIdle:     200,
		MaxActive:   1000,
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
}

// take requires token at a limit per duration for key.
func Allow(key string, duration, limit, requires int) (int64, error) {
	return take(key, duration, limit, requires, pool)
}

// defined limit token created in duration time for key.
// usage: take("key", 60, 600, 1, pool) means limited to 600 per 60s.
func take(key string, duration, limit, requires int, pool *redis.Pool) (int64, error) {
	c := pool.Get()
	defer c.Close()

	var err error
	if err := c.Err(); err != nil {
		return 0, err
	}

	rlScript := redis.NewScript(1, script)
	reply, err := rlScript.Do(c, key, duration, limit, requires)
	if err != nil {
		return 0, err
	}
	return reply.(int64), nil
}
