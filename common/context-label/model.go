package context_label

import "github.com/eolinker/eosc/eocontext"

const (
	LabelModelCompletionTag = "model_completion_tag"

	TagModelCompletion       = "completion"
	TagModelCompletionStream = "completion-stream"
)

func SetModelCompletionTag(ctx eocontext.EoContext) {
	ctx.SetLabel(LabelModelCompletionTag, TagModelCompletion)
}

func SetModelCompletionStreamTag(ctx eocontext.EoContext) {
	ctx.SetLabel(LabelModelCompletionTag, TagModelCompletionStream)
}
