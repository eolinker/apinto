-- Redis操作key
local key = KEYS[1]

-- 扣减次数
local count = tonumber(ARGV[1])

local _count = redis.call('GET',key)

if _count then
	local remain = tonumber(_count)
	if remain < count then
		return -2
	end
	redis.call('INCRBY',key,-count)
	return 0
end
return -1
