package bailianyun

import (
	"encoding/json"
	"fmt"
	ai_convert "github.com/eolinker/apinto/ai-convert"
	"github.com/eolinker/apinto/common/context-label"
	eoscContext "github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"net/url"
	"strings"
	"time"
)

const (
	videoTaskCommitPath = "/services/aigc/video-generation/video-synthesis"
	videoTaskQueryPath  = "/tasks/{task_id}"
)

func init() {
	driverCreate.Set(ai_convert.ModelTypeVideoTaskCommit, func(mt ai_convert.ModelType, c *Config) (ai_convert.IConverterDriver, error) {
		return NewVideoCommit(provider, c.APIKey, c.BaseUrl, mt, time.Minute)
	})
	driverCreate.Set(ai_convert.ModelTypeVideoTaskQuery, func(mt ai_convert.ModelType, c *Config) (ai_convert.IConverterDriver, error) {
		return NewVideoQuery(provider, c.APIKey, c.BaseUrl, mt, time.Minute)
	})
}

var _ ai_convert.IConverterDriver = (*VideoTaskCommit)(nil)

func NewVideoCommit(provider string, apikey string, baseUrl string, modelType ai_convert.ModelType, timeout time.Duration) (ai_convert.IConverterDriver, error) {
	c := &VideoTaskCommit{
		provider:  provider,
		apikey:    apikey,
		modelType: modelType,
		path:      videoTaskCommitPath,
	}

	balanceHandler, err := ai_convert.NewBalanceHandler(apikey, baseUrl, timeout)
	if err != nil {
		return nil, err
	}
	c.balanceHandler = balanceHandler
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	if strings.TrimSuffix(u.Path, "/") == "" {
		c.path = "/api/v1" + c.path
	} else {
		c.path = fmt.Sprintf("%s%s", strings.TrimSuffix(u.Path, "/"), c.path)
	}
	return c, nil
}

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
	context_label.SetBillingMode(ctx, context_label.BillingModeTaskCreate)
	context_label.SetTaskIDSetFunc(ctx, func(ctx eoscContext.EoContext) error {
		httpContext, err := http_service.Assert(ctx)
		if err != nil {
			return err
		}
		body := httpContext.Response().GetBody()
		var resp TaskCommitResponse
		err = json.Unmarshal(body, &resp)
		if err != nil {
			return err
		}
		context_label.SetTaskID(ctx, resp.Output.TaskID)
		return nil
	})
	return nil
}

func (v *VideoTaskCommit) ResponseConvert(ctx eoscContext.EoContext) error {
	return nil
}

var _ ai_convert.IConverterDriver = (*VideoTaskQuery)(nil)

func NewVideoQuery(provider string, apikey string, baseUrl string, modelType ai_convert.ModelType, timeout time.Duration) (ai_convert.IConverterDriver, error) {
	c := &VideoTaskQuery{
		provider:  provider,
		apikey:    apikey,
		modelType: modelType,
		path:      videoTaskQueryPath,
	}

	balanceHandler, err := ai_convert.NewBalanceHandler(apikey, baseUrl, timeout)
	if err != nil {
		return nil, err
	}
	c.balanceHandler = balanceHandler
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	if strings.TrimSuffix(u.Path, "/") == "" {
		c.path = "/api/v1" + c.path
	} else {
		c.path = fmt.Sprintf("%s%s", strings.TrimSuffix(u.Path, "/"), c.path)
	}
	return c, nil
}

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
	context_label.SetBillingMode(ctx, context_label.BillingModeTaskQuery)
	context_label.SetTaskStatusParseFunc(ctx, func(ctx eoscContext.EoContext) (string, error) {
		httpContext, err := http_service.Assert(ctx)
		if err != nil {
			return "", err
		}
		body := httpContext.Response().GetBody()
		var resp TaskQueryResponse
		err = json.Unmarshal(body, &resp)
		if err != nil {
			return "", err
		}
		switch resp.Output.TaskStatus {
		case "SUCCEEDED":
			return context_label.TaskStatusSuccess, nil
		case "FAILED", "UNKNOWN":
			return context_label.TaskStatusFailed, nil
		case "RUNNING":
			return context_label.TaskStatusRunning, nil
		}
		return context_label.TaskStatusRunning, nil
	})
	return nil
}

func (v *VideoTaskQuery) ResponseConvert(ctx eoscContext.EoContext) error {
	//TODO implement me
	panic("implement me")
}

type TaskCommitResponse struct {
	Output struct {
		TaskID string `json:"task_id"`
	} `json:"output"`
}

type TaskQueryResponse struct {
	Output struct {
		TaskID     string `json:"task_id"`
		TaskStatus string `json:"task_status"`
	} `json:"output"`
}
