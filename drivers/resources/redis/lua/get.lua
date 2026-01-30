-- get.lua
-- KEYS[1] = token
-- ARGV[1] = bucket_start_index (只统计 >= 此值的字段)

local key = KEYS[1]
local win_start = tonumber(ARGV[1])

local fields = redis.call('HKEYS', key)
local sum = 0
local to_del = {}

for _, field in ipairs(fields) do
	local idx = tonumber(field)
	if idx == nil or idx < win_start then
		table.insert(to_del, field)
	else
		local val = tonumber(redis.call('HGET', key, field) or "0")
		sum = sum + val
	end
end

if #to_del > 0 then
	redis.call('HDEL', key, unpack(to_del))
end

return sum