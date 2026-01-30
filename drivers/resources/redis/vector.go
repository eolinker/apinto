package redis

import (
	"context"
	"embed"
	"fmt"
	"github.com/eolinker/eosc/log"
	redis "github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

type Vector struct {
	name string

	step int64 // 时间步长（纳秒）
	size int64 // 窗口桶大小（槽数）
	cmd  redis.Cmdable
	// 新增：锁 TTL（秒），建议短于业务时间
	lockTTL time.Duration
}

// 建议把脚本提前加载，得到 SHA1，避免每次 EVAL 都传全文
var (
	compareAndAddLua string // 在 init 或启动时加载
	getLua           string
	addLua           string
	//go:embed lua/compare_and_add.lua
	compareAndAddScript embed.FS
	//go:embed lua/get.lua
	getScript embed.FS
	//go:embed lua/add.lua
	addScript embed.FS
)

func init() {
	script, err := compareAndAddScript.ReadFile("lua/compare_and_add.lua")
	if err != nil {
		panic(err)
	}
	compareAndAddLua = string(script)
	script, err = getScript.ReadFile("lua/get.lua")
	if err != nil {
		panic(err)
	}
	getLua = string(script)
	script, err = addScript.ReadFile("lua/add.lua")
	if err != nil {
		panic(err)
	}
	addLua = string(script)
}

// 根据 step 判断粒度并返回保守的 multiplier（用于 ttl = window_sec * multiplier）
func (v *Vector) getExpireMultiplier() int {
	stepSec := v.step / int64(time.Second)

	switch {
	case stepSec == 1: // 秒级
		return 8 // 8倍 + 裕度，够覆盖抖动
	case stepSec >= 60 && stepSec < 3600: // 分钟级（60s ~ 59min）
		return 6
	case stepSec >= 3600: // 小时级及以上
		return 4
	default:
		return 6 // 默认偏保守
	}
}

// 计算窗口总秒数
func (v *Vector) windowSeconds() int64 {
	return v.size * v.step / int64(time.Second)
}

// 获取最终的 TTL 秒数（可直接传给 Lua 的最后一个 ARGV）
func (v *Vector) getTTLSeconds() int {
	winSec := v.windowSeconds()
	if winSec <= 0 {
		return 3600 // 防止异常
	}
	mult := v.getExpireMultiplier()
	// 额外加一点固定裕度（防止边界情况）
	extra := int64(600) // 10分钟
	if winSec > 3600*24 {
		extra = 86400 // 大窗口再加1天
	}
	return int(winSec*int64(mult) + extra)
}

// CompareAndAdd 去锁版本（推荐）
func (v *Vector) CompareAndAdd(ctx context.Context, key string, threshold, delta int64) (int64, bool) {
	token := fmt.Sprintf("strategy-limiting:%s:%s", v.name, key)
	nowNs := time.Now().UnixNano()
	index := nowNs / v.step
	bucketStart := (index / v.size) * v.size

	args := []interface{}{
		strconv.FormatInt(index, 10),
		strconv.FormatInt(bucketStart, 10),
		strconv.FormatInt(threshold, 10),
		strconv.FormatInt(delta, 10),
		strconv.Itoa(v.getTTLSeconds()),
	}

	// 使用 EVAL 执行
	result, err := v.cmd.Eval(ctx, compareAndAddLua, []string{token}, args...).Result()
	if err != nil {
		log.Errorf("CompareAndAdd lua failed: %v", err)
		return 0, false
	}

	res, ok := result.([]interface{})
	if !ok || len(res) != 2 {
		return 0, false
	}

	current, _ := res[0].(int64) // 或用 redis.Int64(res[0])
	success, _ := res[1].(int64)

	return current, success == 1
}
func (v *Vector) Add(ctx context.Context, key string, delta int64) int64 {
	token := fmt.Sprintf("strategy-limiting:%s:%s", v.name, key)
	nowNs := time.Now().UnixNano()
	index := nowNs / v.step
	bucketStart := (index / v.size) * v.size

	args := []interface{}{
		strconv.FormatInt(index, 10),
		strconv.FormatInt(bucketStart, 10),
		strconv.FormatInt(delta, 10),
		strconv.Itoa(v.getTTLSeconds()),
	}

	result, err := v.cmd.Eval(ctx, addLua, []string{token}, args...).Result()
	if err != nil {
		log.Errorf("Add lua failed: %v", err)
		return 0
	}

	val, ok := result.(int64)
	if !ok {
		return 0
	}
	return val
}

func (v *Vector) Get(ctx context.Context, key string) int64 {
	token := fmt.Sprintf("strategy-limiting:%s:%s", v.name, key)
	nowNs := time.Now().UnixNano()
	index := nowNs / v.step
	bucketStart := (index / v.size) * v.size

	result, err := v.cmd.Eval(ctx, getLua, []string{token},
		strconv.FormatInt(bucketStart, 10)).Result()
	if err != nil {
		return 0
	}

	sum, ok := result.(int64)
	if !ok {
		return 0
	}
	return sum
}

func newVector(name string, uin int64, step int64, cmd redis.Cmdable) *Vector {
	return &Vector{name: name, step: step, cmd: cmd, size: uin / step}
}
