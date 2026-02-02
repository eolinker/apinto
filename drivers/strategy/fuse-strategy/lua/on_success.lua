-- KEYS[1] = baseKey
-- ARGV[1] = recover_threshold (连续成功次数阈值)

local base = KEYS[1]
local state = redis.call('HGET', base, 'state')

if state == 'observe' then
	local sc = redis.call('HINCRBY', base, 'succ_count', 1)
	redis.call('EXPIRE', base, 3600)  -- 整体 1 小时过期

	if sc >= tonumber(ARGV[1]) then
		-- 恢复成功，清空所有熔断相关字段
		redis.call('DEL', base)
		return {'recovered_to_healthy', sc}
	end
	return {'success_in_observe', sc}
end

return {'ignored_success'}