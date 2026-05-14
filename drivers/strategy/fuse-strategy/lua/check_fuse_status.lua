-- KEYS[1] = baseKey
-- ARGV[1] = now_unix_second

local base = KEYS[1]
local now = tonumber(ARGV[1])

local state = redis.call('HGET', base, 'state') or 'healthy'
local expire_at = tonumber(redis.call('HGET', base, 'expire_at') or '0')

if state == 'fusing' then
	if now < expire_at then
		return {'fusing', expire_at - now}
	else
		-- 过期，进入 observe，重置半开计数
		redis.call('HMSET', base,
			'state', 'observe',
			'succ_count', '0',
			'err_count', '0'
		)
		redis.call('HEXPIRE', base, 'succ_count', 60)
		redis.call('HEXPIRE', base, 'err_count', 60)
		return {'observe','enter_observe'}
	end
end

if state == 'observe' then
	return {'observe', 'allow_probe'}
end

return {'healthy', 'allow'}