package ai_prompt

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	ai_convert "github.com/eolinker/apinto/ai-convert"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

type RequestMessage struct {
	Model     string            `json:"model"`
	Messages  []Message         `json:"messages"`
	Variables map[string]string `json:"variables,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type executor struct {
	drivers.WorkerBase
	prompt    string
	required  bool
	variables map[string]bool
}

func (e *executor) DoFilter(ctx eocontext.EoContext, next eocontext.IChain) (err error) {
	// 判断是否是websocket
	return http_context.DoHttpFilter(e, ctx, next)
}

func (e *executor) DoHttpFilter(ctx http_context.IHttpContext, next eocontext.IChain) error {
	body, err := ctx.Proxy().Body().RawBody()
	if err != nil {
		return err
	}
	body, err = genRequestMessage(ctx, body, e.prompt, e.variables, e.required)
	if err != nil {
		result := make(map[string]interface{})
		result["code"] = -1
		result["error"] = err.Error()
		marData, _ := json.Marshal(result)
		ctx.Response().SetBody(marData)
		return err
	}
	ctx.Proxy().Body().SetRaw("application/json", body)

	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

var (
	hashServiceMapping = "service_mapping"
)

func genRequestMessage(ctx http_context.IHttpContext, body []byte, prompt string, variables map[string]bool, required bool) ([]byte, error) {
	baseMsg := eosc.NewBase[RequestMessage](nil)
	err := json.Unmarshal(body, baseMsg)
	if err != nil {
		return nil, err
	}
	model := baseMsg.Config.Model
	provider := ctx.GetLabel("provider")
	if provider != "" {
		// 检查是否配置了service_mapping，若无则跳过
		m, has := customerVar.GetAll(fmt.Sprintf("%s:%s", hashServiceMapping, provider))
		if has {
			model = baseMsg.Config.Model
			if model != "" {
				v, ok := m[model]
				if ok {
					// 若配置了服务映射，则使用映射的值
					model = v
				}
			} else {
				v, ok := m["default"]
				if ok {
					// 若配置了服务映射，model值为空，且有默认值，使用默认值
					model = v
				}
			}
		}
	}
	if model != "" {
		// 当参数值非空时，划分Model参数，格式为{供应商ID}/{模型ID}
		ss := strings.SplitN(model, "/", 2)
		if len(ss) < 2 {
			return nil, errors.New("service mapping error")
		}
		ai_convert.SetAIProvider(ctx, ss[0])
		ai_convert.SetAIModel(ctx, ss[1])
		// 重置Model参数，以便后续使用负载
		baseMsg.Config.Model = ""
	}

	if len(baseMsg.Config.Variables) == 0 && required {
		return nil, errors.New("variables is required")
	}

	for k, v := range variables {
		if _, ok := baseMsg.Config.Variables[k]; !ok && v {
			return nil, fmt.Errorf("variable %s is required", k)
		}
		prompt = strings.Replace(prompt, fmt.Sprintf("{{%s}}", k), baseMsg.Config.Variables[k], -1)
	}
	messages := []Message{
		{
			Role:    "system",
			Content: prompt,
		},
	}
	if prompt != "" {
		messages = append(messages, baseMsg.Config.Messages...)
	} else {
		messages = baseMsg.Config.Messages
	}
	// 重制为空
	baseMsg.Config.Variables = nil
	delete(baseMsg.Append, "variables")
	return json.Marshal(baseMsg)
}

func (e *executor) Destroy() {
}

func (e *executor) Start() error {
	return nil
}

func (e *executor) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {

	return nil
}

func (e *executor) reset(cfg *Config, workers map[eosc.RequireId]eosc.IWorker) error {
	variables := make(map[string]bool)
	required := false

	for _, v := range cfg.Variables {
		variables[v.Key] = v.Require
		if v.Require {
			required = true
		}
	}
	e.variables = variables
	e.required = required
	e.prompt = cfg.Prompt
	return nil
}

func (e *executor) Stop() error {
	return nil
}

func (e *executor) CheckSkill(skill string) bool {
	return http_context.FilterSkillName == skill
}
