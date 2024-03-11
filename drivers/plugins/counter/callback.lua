-- Redis操作key
local key = KEYS[1]

-- 返还次数
local count = tonumber(ARGV[2])

local _count = redis.call('GET',key)
if _count then
    redis.call('INCRBY',key,count)
end