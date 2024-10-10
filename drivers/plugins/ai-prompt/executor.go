package ai_prompt

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

type RequestMessage struct {
	Messages  []Message         `json:"messages"`
	Variables map[string]string `json:"variables"`
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
	body, err = genRequestMessage(body, e.prompt, e.variables, e.required)
	if err != nil {
		return err
	}
	ctx.Proxy().Body().SetRaw("application/json", body)

	if next != nil {
		return next.DoChain(ctx)
	}
	return nil
}

func genRequestMessage(body []byte, prompt string, variables map[string]bool, required bool) ([]byte, error) {
	baseMsg := eosc.NewBase[RequestMessage]()
	err := json.Unmarshal(body, baseMsg)
	if err != nil {
		return nil, err
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
	return json.Marshal(map[string]interface{}{
		"messages": messages,
	})
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
