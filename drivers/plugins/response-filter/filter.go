package response_filter

import (
	"fmt"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	"strings"
)

func removeDuplicateStrings(input []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(input))
	for _, str := range input {
		if str == "" {
			continue
		}
		if _, exists := seen[str]; !exists {
			seen[str] = struct{}{}
			result = append(result, str)
		}
	}

	return result
}

type IFilter interface {
	Filter(ctx http_service.IHttpContext) error
}

func NewBodyWhiteFilter(keys []string) (IFilter, error) {
	rules, err := SafeCompile(removeDuplicateStrings(keys))
	if err != nil {
		return nil, err
	}
	return &BodyWhiteFilter{rules: rules}, nil
}

type BodyWhiteFilter struct {
	rules []CompiledRule
}

func (b *BodyWhiteFilter) Filter(ctx http_service.IHttpContext) error {
	newBody, err := Extract(string(ctx.Response().GetBody()), b.rules)
	if err != nil {
		return fmt.Errorf("failed to generate new body: %v", err)
	}
	ctx.Response().SetBody([]byte(newBody))
	return nil
}

func NewBodyBlackFilter(keys []string) (IFilter, error) {
	es, err := newExprSlice(removeDuplicateStrings(keys))
	if err != nil {
		return nil, err
	}
	return &BodyBlackFilter{es: es}, nil
}

type BodyBlackFilter struct {
	es []jp.Expr
}

func (b *BodyBlackFilter) Filter(ctx http_service.IHttpContext) error {
	body := ctx.Response().GetBody()
	n, err := oj.Parse(body)
	if err != nil {
		return err
	}
	for _, filter := range b.es {
		filter.Del(n)
	}
	body, err = oj.Marshal(n)
	ctx.Response().SetBody(body)
	return nil
}

func NewHeaderWhiteFilter(keys []string) (IFilter, error) {
	return &HeaderWhiteFilter{keys: removeDuplicateStrings(keys)}, nil
}

type HeaderWhiteFilter struct {
	keys []string
}

func (h *HeaderWhiteFilter) Filter(ctx http_service.IHttpContext) error {
	header := ctx.Response().Headers()
	ctx.Response().HeaderReset()
	for _, key := range h.keys {
		if value := header.Get(key); value != "" {
			ctx.Response().SetHeader(key, value)
		}
	}
	return nil
}

func NewHeaderBlackFilter(keys []string) (IFilter, error) {
	return &HeaderBlackFilter{keys: removeDuplicateStrings(keys)}, nil
}

type HeaderBlackFilter struct {
	keys []string
}

func (h *HeaderBlackFilter) Filter(ctx http_service.IHttpContext) error {
	for _, key := range h.keys {
		ctx.Response().DelHeader(key)
	}
	return nil
}

func newExprSlice(rules []string) ([]jp.Expr, error) {
	es := make([]jp.Expr, 0, len(rules))
	for _, filter := range rules {
		key := filter
		if !strings.HasPrefix(key, "$.") {
			key = "$." + key
		}
		expr, err := jp.ParseString(filter)
		if err != nil {
			return nil, err
		}
		es = append(es, expr)
	}
	return es, nil
}
