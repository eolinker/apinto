package pricing

import (
	"context"
	"errors"
)

// ErrInsufficientBalance 表示账户余额不足以完成预扣或实际扣费。
var ErrInsufficientBalance = errors.New("insufficient balance")

// DeductMeta 扣减元数据，用于审计与追踪。
type DeductMeta struct {
	Provider  string
	Resource  string
	RequestID string
	TaskID    string
	Phase     PricingPhase
}

// IBalanceManager 余额管理器接口，提供并发安全的两阶段扣减能力。
//
// 两阶段语义：
//  1. PreDeduct  在请求转发前按预估值预扣，原子性地校验余额并冻结资金；
//  2. Settle     上游响应成功后按实际费用结算，多退少补；
//  3. Rollback   上游失败或被拦截时，全额退还冻结金额。
type IBalanceManager interface {
	PreDeduct(ctx context.Context, accountID string, estimatedCost float64, meta DeductMeta) (freezeID string, err error)
	Settle(ctx context.Context, freezeID string, actualCost float64, meta DeductMeta) error
	Rollback(ctx context.Context, freezeID string) error
	GetBalance(ctx context.Context, accountID string) (float64, error)
}
