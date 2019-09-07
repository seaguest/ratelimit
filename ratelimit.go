package ratelimit

import (
	"github.com/gomodule/redigo/redis"
)

const script = `
-- key
local key = KEYS[1]
-- max number of tokens
local max_permits = 100000000
-- duration, in second
local duration = tonumber(KEYS[2])
-- number of produced tokens per duration
local permits_per_duration = tonumber(KEYS[3])
-- required tokens
local required_permits = tonumber(ARGV[1])

-- current time
local time = redis.call('time')
local now_micros = tonumber(time[1]) * 1000000 + tonumber(time[2])

local next_free_ticket_micros = tonumber(redis.call('hget', key, 'next_free_ticket_micros') or now_micros)

-- current stored tokens
local stored_permits = tonumber(redis.call('hget', key, 'stored_permits') or 0)
-- fill interval per token
local stable_interval_micros = 1000000 * duration / permits_per_duration

-- produce token
if (now_micros > next_free_ticket_micros) then
    local new_permits = (now_micros - next_free_ticket_micros) / stable_interval_micros
    stored_permits = math.min(max_permits, stored_permits + new_permits)
    next_free_ticket_micros = now_micros
end

-- consume token
local moment_available = next_free_ticket_micros
local stored_permits_to_spend = math.min(required_permits, stored_permits)
local fresh_permits = required_permits - stored_permits_to_spend;
local wait_micros = fresh_permits * stable_interval_micros

redis.replicate_commands()
redis.call('expire', key, duration*2)

if moment_available == now_micros then
	redis.call('hset', key, 'next_free_ticket_micros', next_free_ticket_micros + wait_micros)
	redis.call('hset', key, 'stored_permits', stored_permits - stored_permits_to_spend)
end

-- return wait time for available token
return moment_available - now_micros
`

var (
	rlScript *redis.Script
)

func init() {
	rlScript = redis.NewScript(3, script)
}

// defined limit token created in duration time for key.
// usage: take("key", 60, 600, 1, pool) means limited to 600 per 60s.
func Take(key string, duration, limit, requires int, pool *redis.Pool) (int64, error) {
	c := pool.Get()
	defer c.Close()

	var err error
	if err := c.Err(); err != nil {
		return 0, err
	}

	reply, err := rlScript.Do(c, key, duration, limit, requires)
	if err != nil {
		return 0, err
	}
	return reply.(int64), nil
}
