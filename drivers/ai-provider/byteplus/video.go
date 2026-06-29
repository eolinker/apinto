package byteplus

import (
	"encoding/json"
	"fmt"
	ai_convert "github.com/eolinker/apinto/ai-convert"
	context_label "github.com/eolinker/apinto/utils/context-label"
	eoscContext "github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"net/url"
	"strings"
	"time"
)

func init() {
	driverCreate.Set(ai_convert.ModelTypeVideoTaskCommit, func(mt ai_convert.ModelType, c *Config) (ai_convert.IConverterDriver, error) {
		return NewVideoTaskCommit(provider, c.APIKey, c.BaseUrl, mt, 30*time.Second)
	})
	driverCreate.Set(ai_convert.ModelTypeVideoTaskQuery, func(mt ai_convert.ModelType, c *Config) (ai_convert.IConverterDriver, error) {
		return NewVideoTaskQuery(provider, c.APIKey, c.BaseUrl, mt, 30*time.Second)
	})
}

const (
	videoTaskCommitPath = "/contents/generations/tasks"
	videoTaskQueryPath  = "/contents/generations/tasks"
)

var _ ai_convert.IConverterDriver = (*VideoTaskCommit)(nil)

func NewVideoTaskCommit(provider string, apikey string, base string, modelType ai_convert.ModelType, timeout time.Duration) (ai_convert.IConverterDriver, error) {
	c := &VideoTaskCommit{
		provider:  provider,
		apikey:    apikey,
		modelType: modelType,
		path:      videoTaskCommitPath,
	}
	if base == "" {
		base = baseUrl
	}
	balanceHandler, err := ai_convert.NewBalanceHandler(apikey, base, timeout)
	if err != nil {
		return nil, err
	}
	c.balanceHandler = balanceHandler
	u, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	if u.Path != "" {
		c.path = fmt.Sprintf("/%s/%s", strings.Trim(u.Path, "/"), strings.Trim(c.path, "/"))
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
	httpContext.Proxy().Header().SetHeader("Authorization", "Bearer "+v.apikey)
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
		context_label.SetTaskID(ctx, resp.Id)
		return nil
	})
	return nil
}

func (v *VideoTaskCommit) ResponseConvert(ctx eoscContext.EoContext) error {
	return nil
}

var _ ai_convert.IConverterDriver = (*VideoTaskQuery)(nil)

func NewVideoTaskQuery(provider string, apikey string, base string, modelType ai_convert.ModelType, timeout time.Duration) (ai_convert.IConverterDriver, error) {
	c := &VideoTaskQuery{
		provider:  provider,
		apikey:    apikey,
		modelType: modelType,
		path:      videoTaskQueryPath,
	}
	if base == "" {
		base = baseUrl
	}
	balanceHandler, err := ai_convert.NewBalanceHandler(apikey, base, timeout)
	if err != nil {
		return nil, err
	}
	c.balanceHandler = balanceHandler
	u, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	if u.Path != "" {
		c.path = fmt.Sprintf("/%s/%s", strings.Trim(u.Path, "/"), strings.Trim(c.path, "/"))
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
	taskId := context_label.GetTaskID(ctx)
	httpContext.Proxy().Header().SetHeader("Authorization", "Bearer "+v.apikey)
	httpContext.Proxy().URI().SetPath(fmt.Sprintf("%s/%s", v.path, taskId))
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
		switch resp.Status {
		case "succeeded":
			return context_label.TaskStatusSuccess, nil
		case "failed":
			return context_label.TaskStatusFailed, nil
		case "running":
			return context_label.TaskStatusRunning, nil
		}
		return context_label.TaskStatusRunning, nil
	})
	return nil
}

func (v *VideoTaskQuery) ResponseConvert(ctx eoscContext.EoContext) error {
	return nil
}

type TaskCommitResponse struct {
	Id string `json:"id"`
}

type TaskQueryResponse struct {
	Model  string `json:"model"`
	Status string `json:"status"`
}
