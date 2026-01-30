-- compare_and_add.lua
-- KEYS[1]   = hash key (token)
-- ARGV[1]   = current_index (int64 string)
-- ARGV[2]   = bucket_start_index (窗口起始index)
-- ARGV[3]   = threshold (int64 string)
-- ARGV[4]   = delta (int64 string，通常 "1")

local key       = KEYS[1]
local curr_idx  = tonumber(ARGV[1])
local win_start = tonumber(ARGV[2])
local threshold = tonumber(ARGV[3])
local delta     = tonumber(ARGV[4] or "1")

-- 1. 清理过期字段（< win_start）
local fields = redis.call('HKEYS', key)
local to_del = {}
local sum    = 0

for _, field in ipairs(fields) do
	local idx = tonumber(field)
	if idx and idx < win_start then
		table.insert(to_del, field)
	elseif idx and idx >= win_start then
		local val = tonumber(redis.call('HGET', key, field) or "0")
		sum = sum + val
	end
end

-- 2. 判断是否已经超限
if sum >= threshold then
	-- 已超，返回当前 sum 和 false
	if #to_del > 0 then
		redis.call('HDEL', key, unpack(to_del))
	end
	return {sum, 0}
end

-- 3. 没超，增加
local new_sum = sum + delta
redis.call('HINCRBY', key, curr_idx, delta)

-- 4. 清理（如果有的话）
if #to_del > 0 then
	redis.call('HDEL', key, unpack(to_del))
end

-- ARGV[#ARGV] 是 Go 传进来的 ttl 秒数（最后一个参数）
redis.call('EXPIRE', KEYS[1], ARGV[#ARGV])

return {new_sum, 1}