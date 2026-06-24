package context_label

import (
	"github.com/eolinker/eosc/eocontext"
)

const (
	LabelStreamRunning  = "stream_running"
	LabelCurrentRunning = "current_running"
	LabelDisableStream  = "disable_stream"
	LabelStreamJsonBody = "stream_json_body"
)

func IsStreamRunning(ctx eocontext.EoContext) bool {
	value := ctx.GetLabel(LabelStreamRunning)
	return value == "true"
}

func SetStreamRunning(ctx eocontext.EoContext, running bool) {
	value := "false"
	if running {
		value = "true"
	}
	ctx.SetLabel(LabelStreamRunning, value)
}

func IsCurrentRunning(ctx eocontext.EoContext) bool {
	value := ctx.GetLabel(LabelCurrentRunning)
	return value == "true"
}

func SetCurrentRunning(ctx eocontext.EoContext, running bool) {
	value := "false"
	if running {
		value = "true"
	}
	ctx.SetLabel(LabelCurrentRunning, value)
}

func SetDisableStream(ctx eocontext.EoContext, disable bool) {
	value := "false"
	if disable {
		value = "true"
	}
	ctx.SetLabel(LabelDisableStream, value)
}

func IsDisableStream(ctx eocontext.EoContext) bool {
	value := ctx.GetLabel(LabelDisableStream)
	return value == "true"
}
