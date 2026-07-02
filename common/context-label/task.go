package context_label

import "github.com/eolinker/eosc/eocontext"

const (
	LabelTaskID = "task_id"
	//LabelTaskStatus          = "task_status"
	LabelTaskStatusParseFunc = "task_status_parse_func"
	LabelTaskIDSetFunc       = "task_id_set_func"
	TaskStatusRunning        = "running"
	TaskStatusSuccess        = "success"
	TaskStatusFailed         = "failed"
)

type TaskStatusParseFunc func(ctx eocontext.EoContext) (string, error)

type TaskIDSetFunc func(ctx eocontext.EoContext) error

func GetTaskID(ctx eocontext.EoContext) string {
	return ctx.GetLabel(LabelTaskID)
}

func SetTaskID(ctx eocontext.EoContext, taskID string) {
	ctx.SetLabel(LabelTaskID, taskID)
}

func SetTaskStatusParseFunc(ctx eocontext.EoContext, parseFunc TaskStatusParseFunc) {
	ctx.WithValue(LabelTaskStatusParseFunc, parseFunc)
}

func GetTaskStatusParseFunc(ctx eocontext.EoContext) TaskStatusParseFunc {
	value := ctx.Value(LabelTaskStatusParseFunc)
	if value == nil {
		return nil
	}
	parseFunc, ok := value.(TaskStatusParseFunc)
	if !ok {
		return nil
	}
	return parseFunc
}

func SetTaskIDSetFunc(ctx eocontext.EoContext, setFunc TaskIDSetFunc) {
	ctx.WithValue(LabelTaskIDSetFunc, setFunc)
}

func GetTaskIDSetFunc(ctx eocontext.EoContext) TaskIDSetFunc {
	value := ctx.Value(LabelTaskIDSetFunc)
	if value == nil {
		return nil
	}
	setFunc, ok := value.(TaskIDSetFunc)
	if !ok {
		return nil
	}
	return setFunc
}
