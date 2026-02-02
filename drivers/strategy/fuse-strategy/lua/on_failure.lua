-- KEYS[1] = baseKey
-- ARGV[1] = err_threshold     (int)
-- ARGV[2] = base_fuse_seconds (int)
-- ARGV[3] = max_fuse_seconds  (int)
-- ARGV[4] = now_unix_second

local base = KEYS[1]
local now = tonumber(ARGV[4])
local state = redis.call('HGET', base, 'state') or 'healthy'

if state == 'healthy' then
	local err_cnt = redis.call('HINCRBY', base, 'err_count', 1)
	redis.call('EXPIRE', base, 600)  -- 整体 10 分钟过期

	if err_cnt > tonumber(ARGV[1]) then
		local fc = redis.call('HINCRBY', base, 'fuse_count', 1)
		local exp_sec = fc * tonumber(ARGV[2])
		if exp_sec > tonumber(ARGV[3]) then
			exp_sec = tonumber(ARGV[3])
		end

		redis.call('HMSET', base,
			'state', 'fusing',
			'expire_at', now + exp_sec,
			'err_count', '0'
		)
		redis.call('EXPIRE', base, 1200)  -- 熔断后保留 20 分钟
		return {'trigger_fusing', exp_sec}
	end
	return {'counted_error', err_cnt}

elseif state == 'observe' then
	-- 半开期失败 → 立即回熔断，并增加熔断次数
	redis.call('HDEL', base, 'succ_count')
	local fc = tonumber(redis.call('HGET', base, 'fuse_count') or '0') + 1
	redis.call('HSET', base, 'fuse_count', fc)

	local exp_sec = fc * tonumber(ARGV[2])
	if exp_sec > tonumber(ARGV[3]) then
		exp_sec = tonumber(ARGV[3])
	end

	redis.call('HMSET', base,
		'state', 'fusing',
		'expire_at', now + exp_sec
	)
	return {'halfopen_failed_back_to_fusing', exp_sec}
elseif state == 'fusing' then
	return {'ignored_in_fusing'}
end

return {'ignored'}