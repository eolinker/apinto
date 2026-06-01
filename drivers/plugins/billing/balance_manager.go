package billing

import (
	"context"
	"fmt"
	"log"

	"github.com/eolinker/apinto/pricing"
	"github.com/eolinker/apinto/resources"
	scope_manager "github.com/eolinker/apinto/scope-manager"
	"github.com/google/uuid"
)

// RedisBalanceManager 基于 Redis Lua 脚本的分布式账户余额管理。
//
// 余额扣减遵循两阶段模型：
//  1. PreDeduct：原子检查余额并冻结；返回 freezeID 供后续结算/回滚
//  2. Settle：按实际费用结算，多退少补
//  3. Rollback：失败时全额退还冻结金额
//
// Redis 不可用时降级放行（返回 nil + 占位 freezeID），保证主链路可用性。
type RedisBalanceManager struct {
	cache scope_manager.IProxyOutput[resources.ICache]
}

// NewRedisBalanceManager 通过缓存代理构造余额管理器实例。
func NewRedisBalanceManager(cache scope_manager.IProxyOutput[resources.ICache]) *RedisBalanceManager {
	return &RedisBalanceManager{cache: cache}
}

func (r *RedisBalanceManager) getCache() resources.ICache {
	if r == nil || r.cache == nil {
		return nil
	}
	list := r.cache.List()
	if len(list) < 1 {
		return nil
	}
	for _, c := range list {
		if c != nil {
			return c
		}
	}
	return nil
}

// 计费 Redis 键前缀，与旧版 ai-balance/ai-freeze 隔离。
func balanceKey(accountID string) string { return "apinto:billing:balance:" + accountID }
func freezeKey(freezeID string) string   { return "apinto:billing:freeze:" + freezeID }

// PreDeduct 原子预扣费。返回的 freezeID 用于后续结算或回滚。
// Redis 不可用时返回占位 freezeID 并放行，保证主链路可用。
func (r *RedisBalanceManager) PreDeduct(ctx context.Context, accountID string, estimatedCost float64, meta pricing.DeductMeta) (string, error) {
	cache := r.getCache()
	if cache == nil {
		log.Printf("[billing/balance] redis unavailable, bypass PreDeduct for account=%s", accountID)
		return uuid.NewString(), nil
	}

	freezeID := uuid.NewString()

	// Lua: 原子地校验余额并扣减，同时记录冻结条目（HMSET + EXPIRE）
	// KEYS[1]=balanceKey  KEYS[2]=freezeKey
	// ARGV[1]=cost        ARGV[2]=accountID
	script := `
	local balance = tonumber(redis.call('GET', KEYS[1]))
	if balance == nil then
		return -1
	end
	local cost = tonumber(ARGV[1])
	if balance < cost then
		return -2
	end
	redis.call('SET', KEYS[1], tostring(balance - cost))
	redis.call('HMSET', KEYS[2], 'account_id', ARGV[2], 'estimated_cost', ARGV[1])
	redis.call('EXPIRE', KEYS[2], 86400)
	return 1
	`

	res, err := cache.Run(ctx, script, []string{balanceKey(accountID), freezeKey(freezeID)},
		fmt.Sprintf("%f", estimatedCost), accountID).Result()
	if err != nil {
		log.Printf("[billing/balance] PreDeduct lua error: %v", err)
		return "", err
	}

	val, ok := toInt64(res)
	if !ok {
		log.Printf("[billing/balance] invalid lua return type: %T", res)
		return freezeID, nil // 兜底放行
	}
	if val == -1 || val == -2 {
		return "", pricing.ErrInsufficientBalance
	}

	log.Printf("[billing/balance] PreDeduct ok: account=%s cost=%f freezeID=%s", accountID, estimatedCost, freezeID)
	return freezeID, nil
}

// Settle 多退少补地结算实际费用，并删除冻结记录。
func (r *RedisBalanceManager) Settle(ctx context.Context, freezeID string, actualCost float64, meta pricing.DeductMeta) error {
	cache := r.getCache()
	if cache == nil {
		log.Printf("[billing/balance] redis unavailable, bypass Settle for freezeID=%s", freezeID)
		return nil
	}

	// Lua: 读取冻结记录差额回写 balance，最后删除冻结记录
	// KEYS[1]=freezeKey  KEYS[2]=balance 前缀
	// ARGV[1]=actualCost
	script := `
	local account_id = redis.call('HGET', KEYS[1], 'account_id')
	local estimated_cost = tonumber(redis.call('HGET', KEYS[1], 'estimated_cost'))
	if not account_id or estimated_cost == nil then
		return -1
	end
	local actual_cost = tonumber(ARGV[1])
	local diff = estimated_cost - actual_cost
	local balance_key = KEYS[2] .. account_id
	if diff ~= 0 then
		local balance = tonumber(redis.call('GET', balance_key))
		if balance == nil then balance = 0 end
		redis.call('SET', balance_key, tostring(balance + diff))
	end
	redis.call('DEL', KEYS[1])
	return 1
	`

	_, err := cache.Run(ctx, script, []string{freezeKey(freezeID), "apinto:billing:balance:"},
		fmt.Sprintf("%f", actualCost)).Result()
	if err != nil {
		log.Printf("[billing/balance] Settle lua error: %v", err)
		return err
	}
	log.Printf("[billing/balance] Settle ok: freezeID=%s actualCost=%f", freezeID, actualCost)
	return nil
}

// Rollback 全额退还冻结金额并删除冻结记录。
func (r *RedisBalanceManager) Rollback(ctx context.Context, freezeID string) error {
	cache := r.getCache()
	if cache == nil {
		log.Printf("[billing/balance] redis unavailable, bypass Rollback for freezeID=%s", freezeID)
		return nil
	}

	script := `
	local account_id = redis.call('HGET', KEYS[1], 'account_id')
	local estimated_cost = tonumber(redis.call('HGET', KEYS[1], 'estimated_cost'))
	if not account_id or estimated_cost == nil then
		return -1
	end
	local balance_key = KEYS[2] .. account_id
	local balance = tonumber(redis.call('GET', balance_key))
	if balance == nil then balance = 0 end
	redis.call('SET', balance_key, tostring(balance + estimated_cost))
	redis.call('DEL', KEYS[1])
	return 1
	`

	_, err := cache.Run(ctx, script, []string{freezeKey(freezeID), "apinto:billing:balance:"}).Result()
	if err != nil {
		log.Printf("[billing/balance] Rollback lua error: %v", err)
		return err
	}
	log.Printf("[billing/balance] Rollback ok: freezeID=%s", freezeID)
	return nil
}

// GetBalance 查询账户当前余额。
func (r *RedisBalanceManager) GetBalance(ctx context.Context, accountID string) (float64, error) {
	cache := r.getCache()
	if cache == nil {
		return 0, nil
	}
	res, err := cache.Get(ctx, balanceKey(accountID)).Result()
	if err != nil {
		return 0, err
	}
	var balance float64
	fmt.Sscanf(res, "%f", &balance)
	return balance, nil
}

func toInt64(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int64:
		return n, true
	case float64:
		return int64(n), true
	default:
		return 0, false
	}
}
