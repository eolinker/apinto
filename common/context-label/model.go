package context_label

import "github.com/eolinker/eosc/eocontext"

const (
	LabelModelCompletionTag = "model_completion_tag"
	LabelPreInputToken      = "pre_input_token"
	LabelTotalToken         = "total_token"
	
	TagModelCompletion       = "completion"
	TagModelCompletionStream = "completion-stream"
)

func SetModelCompletionTag(ctx eocontext.EoContext) {
	ctx.SetLabel(LabelModelCompletionTag, TagModelCompletion)
}

func SetModelCompletionStreamTag(ctx eocontext.EoContext) {
	ctx.SetLabel(LabelModelCompletionTag, TagModelCompletionStream)
}

// --- Token 预扣与实际消费存取函数 ---
func SetPreInputToken(ctx eocontext.EoContext, token int) {
	ctx.WithValue(LabelPreInputToken, token)
}

func GetPreInputToken(ctx eocontext.EoContext) int64 {
	if v := ctx.Value(LabelPreInputToken); v != nil {
		if token, ok := v.(int); ok {
			return int64(token)
		}
		if token, ok := v.(int64); ok {
			return token
		}
	}
	return 0
}

// SetActualTotalToken 设置实际消费的总 Token 数
func SetActualTotalToken(ctx eocontext.EoContext, token int) {
	ctx.WithValue(LabelTotalToken, token)
	SetPriceVariable(ctx, "total_token", token)
}

// GetActualTotalToken 获取实际消费的总 Token 数
func GetActualTotalToken(ctx eocontext.EoContext) int64 {
	if v := ctx.Value(LabelTotalToken); v != nil {
		if token, ok := v.(int); ok {
			return int64(token)
		}
		if token, ok := v.(int64); ok {
			return token
		}
	}
	if v := GetPriceVariable(ctx, "total_token"); v != nil {
		if token, ok := v.(int); ok {
			return int64(token)
		}
		if token, ok := v.(int64); ok {
			return token
		}
	}
	return 0
}
