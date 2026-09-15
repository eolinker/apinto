package context_label

import (
	"errors"
	"time"

	"github.com/eolinker/eosc/eocontext"
)

const (
	LabelAITimeout      = "ai_timeout"
	LabelAITimeoutError = "ai_timeout_error"
	LabelUpstreamCost   = "upstream_cost"
)

var (
	ErrAITimeout = errors.New("upstream request timeout")
)

// SetAITimeout 在上下文设置 AI 上游超时标签
func SetAITimeout(ctx eocontext.EoContext, timeout bool) {
	if timeout {
		ctx.SetLabel(LabelAITimeout, "true")
	} else {
		ctx.SetLabel(LabelAITimeout, "false")
	}
}

// IsAITimeout 检查当前请求是否发生了 AI 上游超时
func IsAITimeout(ctx eocontext.EoContext) bool {
	return ctx.GetLabel(LabelAITimeout) == "true"
}

// HasAITimeoutLabel 检查上下文中是否显式设置了超时标签
func HasAITimeoutLabel(ctx eocontext.EoContext) bool {
	return ctx.GetLabel(LabelAITimeout) != ""
}

// SetAITimeoutError 在上下文中设置超时错误信息以及超时标签
func SetAITimeoutError(ctx eocontext.EoContext, err error) {
	ctx.SetLabel(LabelAITimeout, "true")
	if err != nil {
		ctx.WithValue(LabelAITimeoutError, err)
	}
}

// ClearAITimeout 清除上下文中的超时标记与超时错误
func ClearAITimeout(ctx eocontext.EoContext) {
	ctx.SetLabel(LabelAITimeout, "false")
	ctx.WithValue(LabelAITimeoutError, nil)
}

// GetAITimeoutError 获取上下文中的超时错误信息
func GetAITimeoutError(ctx eocontext.EoContext) error {
	if v := ctx.Value(LabelAITimeoutError); v != nil {
		if err, ok := v.(error); ok {
			return err
		}
		if s, ok := v.(string); ok && s != "" {
			return errors.New(s)
		}
	}
	if v := ctx.Value("ai_timeout_error"); v != nil {
		if err, ok := v.(error); ok {
			return err
		}
		if s, ok := v.(string); ok && s != "" {
			return errors.New(s)
		}
	}
	return nil
}

// SetUpstreamCost 设置网关向上游请求的真实网络耗时
func SetUpstreamCost(ctx eocontext.EoContext, cost time.Duration) {
	ctx.WithValue(LabelUpstreamCost, cost)
}

// GetUpstreamCost 获取网关向上游请求的真实网络耗时
func GetUpstreamCost(ctx eocontext.EoContext) (time.Duration, bool) {
	if v := ctx.Value(LabelUpstreamCost); v != nil {
		if cost, ok := v.(time.Duration); ok {
			return cost, true
		}
	}
	return 0, false
}

