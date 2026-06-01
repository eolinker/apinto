package bailianyun

import (
	ai_convert "github.com/eolinker/apinto/ai-convert"
	eoscContext "github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

const (
	videoTaskCommitPath = "/api/v1/services/aigc/video-generation/video-synthesis"
	videoTaskQueryPath  = "/api/v1/tasks/{task_id}"
)

var _ ai_convert.IConverterDriver = (*VideoTaskCommit)(nil)

type VideoTaskCommit struct {
	provider       string
	apikey         string
	path           string
	modelType      ai_convert.ModelType
	balanceHandler eoscContext.BalanceHandler
}

func (v *VideoTaskCommit) Provider() string {
	return v.provider
}

func (v *VideoTaskCommit) ModelType() ai_convert.ModelType {
	return v.modelType
}

func (v *VideoTaskCommit) RequestConvert(ctx eoscContext.EoContext, extender map[string]interface{}) error {
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	httpContext.Proxy().Header().SetHeader("X-DashScope-Async", "enable")
	httpContext.Proxy().Header().SetHeader("Authorization", v.apikey)
	httpContext.Proxy().URI().SetPath(v.path)
	if v.balanceHandler != nil {
		ctx.SetBalance(v.balanceHandler)
	}
	return nil
}

func (v *VideoTaskCommit) ResponseConvert(ctx eoscContext.EoContext) error {
	//TODO implement me
	panic("implement me")
}

var _ ai_convert.IConverterDriver = (*VideoTaskQuery)(nil)

type VideoTaskQuery struct {
	provider       string
	apikey         string
	path           string
	modelType      ai_convert.ModelType
	balanceHandler eoscContext.BalanceHandler
}

func (v *VideoTaskQuery) Provider() string {
	return v.provider
}

func (v *VideoTaskQuery) ModelType() ai_convert.ModelType {
	return v.modelType
}

func (v *VideoTaskQuery) RequestConvert(ctx eoscContext.EoContext, extender map[string]interface{}) error {
	httpContext, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	// TODO: 提取task_id，判断是否已经存在，如果存在，则直接返回，不转发
	httpContext.Proxy().Header().SetHeader("Authorization", v.apikey)
	httpContext.Proxy().URI().SetPath(v.path)
	if v.balanceHandler != nil {
		ctx.SetBalance(v.balanceHandler)
	}
	return nil
}

func (v *VideoTaskQuery) ResponseConvert(ctx eoscContext.EoContext) error {
	//TODO implement me
	panic("implement me")
}
