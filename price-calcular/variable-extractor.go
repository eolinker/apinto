package price_calcular

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/eolinker/apinto/common/context-label"
	"regexp"
	"strconv"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/eolinker/eosc/log"

	eoscContext "github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/tidwall/gjson"
)

type Variables map[string]*Variable

func (vs Variables) Check() error {
	for name, v := range vs {
		if err := v.Check(name); err != nil {
			return err
		}
	}
	return nil
}

type Variable struct {
	Source string `json:"source" label:"来源" enum:"request_body,response_body,response_status,expression" default:"response_body"`
	Type   string `json:"type" label:"类型" enum:"string,integer,float,boolean,array" default:"integer"`
	Path   string `json:"path" label:"Json Path 或表达式（source 为 expression 时填写算术表达式）"`
}

func (v *Variable) Check(name string) error {
	if name == "" {
		return fmt.Errorf("variable name cannot be empty")
	}
	switch v.Source {
	case "request_body", "response_body":
		if v.Path == "" {
			return fmt.Errorf("path cannot be empty for v %s with source %s", name, v.Source)
		}
	case "response_status":
		// 状态码提取器不需要配置 json path
	case "expression":
		if v.Path == "" {
			return fmt.Errorf("expression cannot be empty for v %s with source expression", name)
		}
		// 提前校验表达式语法
		if _, err := govaluate.NewEvaluableExpression(preProcessExpression(v.Path)); err != nil {
			return fmt.Errorf("invalid expression '%s' for v %s: %w", v.Path, name, err)
		}
	default:
		return fmt.Errorf("unsupported source %s for v %s", v.Source, name)
	}

	switch v.Type {
	case "string", "integer", "float", "boolean", "array":
		// 合法的目标类型
	default:
		return fmt.Errorf("unsupported type %s for v %s", v.Type, name)
	}
	return nil
}

type IExtractor interface {
	ExtractAll(ctx eoscContext.EoContext) map[string]interface{}
	ExtractAllFromChunk(ctx eoscContext.EoContext, chunk []byte) map[string]interface{}
}

// IVariableExtractor 定义了从普通请求上下文中提取自定义变量的统一接口。
type IVariableExtractor interface {
	// Extract 从 EoContext 提取出变量的具体值，并在提取后自动将其转为目标类型。
	Extract(ctx eoscContext.EoContext) (interface{}, error)
}

// IStreamVariableExtractor 定义了从流式响应的一个数据包 Chunk (字节切片) 中，
// 逐帧定位、解包并提取自定义变量值的统一接口。
type IStreamVariableExtractor interface {
	// ExtractFromChunk 从 []byte 响应数据块中解析并提取特定变量的目标类型值。
	ExtractFromChunk(chunk []byte) (interface{}, error)
}

// BodyExtractor 负责从 HTTP 请求体 (request_body) 或响应体 (response_body)
// 中通过指定的 JSON Path (gjson 语法) 极速提取指定的变量，无需完整的反序列化，性能优异。
// 同时实现了 IVariableExtractor 与 IStreamVariableExtractor 接口。
type BodyExtractor struct {
	path    string // JSON Path 表达式路径
	isReq   bool   // 是否是请求体（true 为 request_body，false 为 response_body）
	varType string // 目标数据类型（如 string, integer, float, boolean）
}

var jsonPathRegex = regexp.MustCompile(`^(\$?\.?)(.*?)\[\?\(@\.([a-zA-Z0-9_-]+)\s*([=!<>]+)\s*(.*?)\)\](.*)$`)

func convertJSONPathToGjson(path string) string {
	if jsonPathRegex.MatchString(path) {
		path = jsonPathRegex.ReplaceAllString(path, "${2}.#(${3}${4}${5})${6}")
	}
	// Also strip leading $. if present
	if strings.HasPrefix(path, "$.") {
		path = path[2:]
	} else if strings.HasPrefix(path, "$") {
		path = path[1:]
	}
	return path
}

// NewBodyExtractor 实例化并配置一个 Body 提取器。
func NewBodyExtractor(path string, isReq bool, varType string) (*BodyExtractor, error) {
	return &BodyExtractor{
		path:    convertJSONPathToGjson(path),
		isReq:   isReq,
		varType: varType,
	}, nil
}

// Extract 使用 tidwall/gjson 高性能引擎从普通 HTTP 请求/响应体中提取变量。
func (e *BodyExtractor) Extract(ctx eoscContext.EoContext) (interface{}, error) {
	httpCtx, err := http_context.Assert(ctx)
	if err != nil {
		return nil, err
	}
	var body []byte
	if e.isReq {
		body, _ = httpCtx.Proxy().Body().RawBody()
	} else {
		body = httpCtx.Response().GetBody()
	}
	if len(body) == 0 {
		return nil, errors.New("empty body")
	}

	// 极速单点查询，不耗费完整序列化开销
	res := gjson.GetBytes(body, e.path)
	//if !res.Exists() {
	//	return nil, fmt.Errorf("gjson path '%s' matches no values", e.path)
	//}

	return convertGjsonType(res, e.varType)
}

// ExtractFromChunk 实现了 IStreamVariableExtractor，
// 支持从流式（Server-Sent Events，即 SSE 协议等）响应的 chunk 原始字节块中提取并转换目标变量。
func (e *BodyExtractor) ExtractFromChunk(chunk []byte) (interface{}, error) {
	if len(chunk) == 0 {
		if e.varType == "boolean" {
			return false, nil
		}
		if e.varType == "array" {
			return []interface{}{}, nil
		}
		return nil, errors.New("empty chunk")
	}

	// 将 chunk 换行分割，支持单次 TCP 封包（TCP粘包）合并传输多帧 SSE 数据的解析
	lines := bytes.Split(chunk, []byte("\n"))
	var lastVal interface{}
	var hasValue bool

	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// 解析 SSE 协议规范中以 "data: " 标签开头的有效载荷帧
		if bytes.HasPrefix(line, []byte("data: ")) {
			data := bytes.TrimPrefix(line, []byte("data: "))
			dataStr := string(bytes.TrimSpace(data))

			// 忽略流结束标志
			if dataStr == "[DONE]" {
				continue
			}

			// 在 SSE 载荷帧内，执行指定的 json path 查询
			res := gjson.Get(dataStr, e.path)
			//if !res.Exists() {
			//	continue
			//}

			val, err := convertGjsonType(res, e.varType)
			if err == nil {
				lastVal = val
				hasValue = true
			}
		}
	}

	if !hasValue {
		// 降级保护：如果整体没有找到 "data: " 帧，尝试将整个 chunk 作为普通 JSON 包直接定位提取
		res := gjson.GetBytes(chunk, e.path)
		if res.Exists() {
			return convertGjsonType(res, e.varType)
		}
		if e.varType == "boolean" {
			return false, nil
		}
		if e.varType == "array" {
			return []interface{}{}, nil
		}
		return nil, fmt.Errorf("gjson path '%s' matches no values in chunk", e.path)
	}

	return lastVal, nil
}

// StatusExtractor 负责从 HTTP 响应状态码 (response_status) 中提取变量。
type StatusExtractor struct {
	varType string // 目标类型
}

// NewStatusExtractor 实例化一个响应状态码提取器。
func NewStatusExtractor(varType string) *StatusExtractor {
	return &StatusExtractor{varType: varType}
}

// Extract 从上下文中提取状态码并执行类型转换。
func (e *StatusExtractor) Extract(ctx eoscContext.EoContext) (interface{}, error) {
	httpCtx, err := http_context.Assert(ctx)
	if err != nil {
		return nil, err
	}
	statusCode := httpCtx.Response().StatusCode()
	return convertRawType(statusCode, e.varType)
}

// ExpressionExtractor 负责基于其它已提取出的变量，通过算术表达式（govaluate）计算派生变量。
// 例如 total_token = input_token + output_token。
// 它不直接从上下文提取，而是依赖 VariablesExtractor 在常规变量提取完成后调用。
type ExpressionExtractor struct {
	expr    *govaluate.EvaluableExpression // 预编译后的算术表达式
	varType string                         // 目标类型
}

// NewExpressionExtractor 实例化一个表达式派生变量提取器。
func NewExpressionExtractor(expression, varType string) (*ExpressionExtractor, error) {
	expr, err := govaluate.NewEvaluableExpression(preProcessExpression(expression))
	if err != nil {
		return nil, err
	}
	return &ExpressionExtractor{expr: expr, varType: varType}, nil
}

// ErrDependencyNotReady 表示表达式依赖的某个变量尚未被提取到（常见于流式早期帧），
// 调用方应据此跳过本次输出，而非视为真正的错误。
var ErrDependencyNotReady = errors.New("dependency v not yet extracted")

// EvaluateFromVars 基于已提取的变量集求值表达式，并转换为目标类型。
// 当表达式引用的任一变量在 vars 中尚未提取到时，返回 ErrDependencyNotReady 以跳过本次求值，
// 避免在流式场景中因依赖变量还未出现而提前写出错误（如 0）的派生值。
func (e *ExpressionExtractor) EvaluateFromVars(vars map[string]interface{}) (interface{}, error) {
	params := make(map[string]interface{}, len(e.expr.Vars()))
	for _, v := range e.expr.Vars() {
		val, ok := vars[v]
		if !ok {
			return nil, ErrDependencyNotReady
		}
		params[v] = val
	}
	result, err := e.expr.Evaluate(params)
	if err != nil {
		return nil, err
	}
	return convertRawType(result, e.varType)
}

// convertGjsonType 将 gjson.Result 结果转换为配置所需的目标强类型。
func convertGjsonType(res gjson.Result, targetType string) (interface{}, error) {
	switch targetType {
	case "string":
		return res.String(), nil
	case "integer":
		return res.Int(), nil
	case "float":
		return res.Float(), nil
	case "boolean":
		if !res.Exists() {
			return false, nil
		}
		if res.Type == gjson.True {
			return true, nil
		}
		if res.Type == gjson.False {
			return false, nil
		}
		if res.Type == gjson.Null {
			return false, nil
		}
		if res.IsObject() || res.IsArray() {
			return true, nil
		}
		return res.Bool(), nil
	case "array":
		if res.IsArray() {
			arr := res.Array()
			vals := make([]interface{}, 0, len(arr))
			for _, item := range arr {
				vals = append(vals, item.Value())
			}
			return vals, nil
		}
		return []interface{}{res.Value()}, nil
	default:
		return res.Value(), nil
	}
}

// convertRawType 负责将原生数值等通用对象动态转为目标类型。
func convertRawType(val interface{}, targetType string) (interface{}, error) {
	if val == nil {
		if targetType == "boolean" {
			return false, nil
		}
		if targetType == "array" {
			return []interface{}{}, nil
		}
		return nil, errors.New("value is nil")
	}
	strVal := fmt.Sprintf("%v", val)
	switch targetType {
	case "string":
		return strVal, nil
	case "integer":
		if f, err := strconv.ParseFloat(strVal, 64); err == nil {
			return int64(f), nil
		}
		i, err := strconv.ParseInt(strVal, 10, 64)
		if err != nil {
			return nil, err
		}
		return i, nil
	case "float":
		f, err := strconv.ParseFloat(strVal, 64)
		if err != nil {
			return nil, err
		}
		return f, nil
	case "boolean":
		b, err := strconv.ParseBool(strVal)
		if err != nil {
			return false, nil
		}
		return b, nil
	case "array":
		return []interface{}{val}, nil
	default:
		return val, nil
	}
}

// VariablesExtractor 是一个聚合提取器，内部包含配置中指定的所有自定义变量的提取逻辑。
type VariablesExtractor struct {
	extractors     map[string]IVariableExtractor   // 变量名称与具体提取器的映射表
	exprExtractors map[string]*ExpressionExtractor // 表达式派生变量（依赖其它变量计算）
}

// NewVariablesExtractor 根据配置中的变量定义字典，初始化一个多变量提取调度器。
func NewVariablesExtractor(vars map[string]*Variable) (IExtractor, error) {
	if len(vars) < 1 {
		return nil, nil
	}
	extractors := make(map[string]IVariableExtractor, len(vars))
	exprExtractors := make(map[string]*ExpressionExtractor)
	for name, v := range vars {
		switch v.Source {
		case "request_body":
			ex, err := NewBodyExtractor(v.Path, true, v.Type)
			if err != nil {
				return nil, fmt.Errorf("v %s error: %w", name, err)
			}
			extractors[name] = ex
		case "response_body":
			ex, err := NewBodyExtractor(v.Path, false, v.Type)
			if err != nil {
				return nil, fmt.Errorf("v %s error: %w", name, err)
			}
			extractors[name] = ex
		case "response_status":
			extractors[name] = NewStatusExtractor(v.Type)
		case "expression":
			ex, err := NewExpressionExtractor(v.Path, v.Type)
			if err != nil {
				return nil, fmt.Errorf("v %s error: %w", name, err)
			}
			exprExtractors[name] = ex
		default:
			return nil, fmt.Errorf("unsupported source %s for v %s", v.Source, name)
		}
	}
	return &VariablesExtractor{extractors: extractors, exprExtractors: exprExtractors}, nil
}

// evalExpressions 基于已提取的常规变量集，计算所有表达式派生变量并写回结果集与上下文。
func (ve *VariablesExtractor) evalExpressions(ctx eoscContext.EoContext, res map[string]interface{}) {
	for name, ex := range ve.exprExtractors {
		val, err := ex.EvaluateFromVars(res)
		if err != nil {
			// 依赖未就绪属于正常情况（如流式早期帧），静默跳过不输出
			if !errors.Is(err, ErrDependencyNotReady) {
				log.Errorf("expression v extractor error for %s: %v", name, err)
			}
			continue
		}
		res[name] = val
		context_label.SetPriceVariable(ctx, name, val)
	}
}

// ExtractAll 驱动执行所有已配置的提取器，返回包含所有提取成功变量的 map 数据集。
func (ve *VariablesExtractor) ExtractAll(ctx eoscContext.EoContext) map[string]interface{} {
	res := make(map[string]interface{}, len(ve.extractors))
	for name, ex := range ve.extractors {
		val, err := ex.Extract(ctx)
		if err == nil {
			res[name] = val
			context_label.SetPriceVariable(ctx, name, val)
		} else {

		}
	}
	// 常规变量提取完成后，再计算依赖它们的表达式派生变量
	ve.evalExpressions(ctx, res)
	return res
}

// ExtractAllFromChunk 驱动执行所有已配置且支持流式提取的提取器，从当前流式 Chunk 里抓取所有成功提取的变量。
func (ve *VariablesExtractor) ExtractAllFromChunk(ctx eoscContext.EoContext, chunk []byte) map[string]interface{} {
	res := make(map[string]interface{}, len(ve.extractors))
	for name, ex := range ve.extractors {
		if streamEx, ok := ex.(IStreamVariableExtractor); ok {
			val, err := streamEx.ExtractFromChunk(chunk)
			if err == nil {
				res[name] = val
				context_label.SetPriceVariable(ctx, name, val)
			} else {
				log.Errorf("stream v extractor error for %s: %v", name, err)
			}
		}
	}
	// 基于本 chunk 提取出的常规变量与上下文已累积的变量，计算表达式派生变量
	if len(ve.exprExtractors) > 0 {
		merged := context_label.GetPriceVariables(ctx)
		if merged == nil {
			merged = make(map[string]interface{}, len(res))
		}
		for k, v := range res {
			merged[k] = v
		}
		for name, ex := range ve.exprExtractors {
			val, err := ex.EvaluateFromVars(merged)
			if err != nil {
				// 依赖未就绪属于正常情况（如流式早期帧），静默跳过不输出
				if !errors.Is(err, ErrDependencyNotReady) {
					log.Errorf("expression v extractor error for %s: %v", name, err)
				}
				continue
			}
			res[name] = val
			context_label.SetPriceVariable(ctx, name, val)
		}
	}
	return res
}
