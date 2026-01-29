-- add.lua
-- KEYS[1] = token
-- ARGV[1] = current_index
-- ARGV[2] = bucket_start (用于清理)
-- ARGV[3] = delta

local key = KEYS[1]
local curr_idx = tonumber(ARGV[1])
local win_start = tonumber(ARGV[2])
local delta = tonumber(ARGV[3] or "1")

-- 清理过期字段
local fields = redis.call('HKEYS', key)
local to_del = {}

for _, field in ipairs(fields) do
	local idx = tonumber(field)
	if idx and idx < win_start then
		table.insert(to_del, field)
	end
end

-- 增加
local new_val = redis.call('HINCRBY', key, curr_idx, delta)

-- 清理
if #to_del > 0 then
	redis.call('HDEL', key, unpack(to_del))
end

-- 设置过期（粗略，2倍窗口 + 裕度）
redis.call('EXPIRE', key, 2 * (curr_idx - win_start + 100))

return new_val