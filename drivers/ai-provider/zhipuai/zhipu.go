package zhipu

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/eolinker/eosc/log"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

func init() {
	ai_convert.RegisterConverterCreateFunc("zhipuai", Create)
}

type Config struct {
	APIKey  string `json:"zhipuai_api_key"`
	BaseUrl string `json:"base_url"`
}

// checkConfig validates the provided configuration.
// It ensures the required fields are set and checks the validity of the Base URL if provided.
//
// Parameters:
//   - v: An interface{} expected to be a pointer to a Config struct.
//
// Returns:
//   - *Config: The validated configuration cast to *Config.
//   - error: An error if the validation fails, or nil if it succeeds.
func checkConfig(conf *Config) error {
	// Check if the APIKey is provided. It is a required field.
	if conf.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}
	if conf.BaseUrl != "" {
		u, err := url.Parse(conf.BaseUrl)
		if err != nil {
			// Return an error if the Base URL cannot be parsed.
			return fmt.Errorf("base url is invalid")
		}
		// Ensure the parsed URL contains both a scheme and a host.
		if u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("base url is invalid")
		}
	}
	return nil
}

func Create(cfg string) (ai_convert.IConverter, error) {
	var conf Config
	err := json.Unmarshal([]byte(cfg), &conf)
	if err != nil {
		return nil, err
	}
	err = checkConfig(&conf)
	if err != nil {
		return nil, err
	}

	return ai_convert.NewOpenAIConvert(conf.APIKey, conf.BaseUrl, 0, errorCallback)
}

func errorCallback(ctx http_service.IHttpContext, body []byte) {
	// http status code
	// 200 - 业务处理成功
	// 400 - 参数错误或文件内容异常
	// 401 - 认证失败或Token超时
	// 404 - 微调功能不可用或微调任务不存在
	// 429 - 接口请求并发超限、文件上传频率过快、账户余额耗尽、账户异常或终端账户异常
	// 434 - 无API权限，微调API和文件管理API处于Beta阶段
	// 435 - 文件大小超过100MB
	// 500 - 服务器在处理请求时发生错误
	switch ctx.Response().StatusCode() {

	case 429:
		// 业务错误码汇总
		// 基本错误
		// 500 - 内部错误
		// 认证错误
		// 1000 - 认证失败
		// 1001 - Header中未接收到认证参数，无法认证
		// 1002 - 无效的认证Token，请确认认证Token的正确传递
		// 1003 - 认证Token已过期，请重新生成/获取
		// 1004 - 提供的认证Token认证失败
		// 账户错误
		// 1100 - 账户读写
		// 1110 - 账户当前处于非活跃状态，请检查账户信息
		// 1111 - 账户不存在
		// 1112 - 账户已被锁定，请与客服联系解锁
		// 1113 - 账户欠费，请充值后重试
		// 1120 - 无法成功访问账户，请稍后再试
		// API调用错误
		// 1200 - API调用错误
		// 1210 - API调用参数不正确，请检查文档
		// 1211 - 模型不存在，请检查模型代码
		// 1212 - 当前模型不支持${method}调用方法
		// 1213 - ${field}参数未正确接收
		// 1214 - ${field}参数无效，请检查文档
		// 1215 - ${field1}和${field2}不能同时设置，请检查文档
		// 1220 - 您没有权限访问${API_name}
		// 1221 - API ${API_name}已下线
		// 1222 - API ${API_name}不存在
		// 1230 - API调用过程错误
		// 1231 - 您已有请求：${request_id}
		// 1232 - 获取异步请求结果时，请使用task_id
		// 1233 - 任务：${task_id}不存在
		// 1234 - 网络错误，错误id：${error_id}，请与客服联系
		// 1235 - 网络错误，错误id：${error_id}，请与客服联系
		// 1260 - API运行时错误
		// 1261 - 提示过长
		// API策略阻断错误
		// 1300 - API调用被策略阻断
		// 1301 - 系统检测到输入或生成中可能存在不安全或敏感内容，请避免使用可能生成敏感内容的提示
		// 1302 - 此API的高并发使用，请降低并发或联系客服提高限制
		// 1303 - 此API的高频率使用，请降低频率或联系客服提高限制
		// 1304 - 此API的日调用限额已达到，如需更多请求，请与客服联系购买
		// 1305 - 目前API请求过多，请稍后再试
		var data ai_convert.Response
		err := json.Unmarshal(body, &data)
		if err != nil {
			log.Errorf("unmarshal body error: %v, body: %s", err, body)
			return
		}
		switch data.Error.Code {
		case "1001", "1002", "1003", "1004":
			// Handle the auth error.
			ai_convert.SetAIStatusInvalid(ctx)
		case "1110", "1111", "1112", "1113", "1120":
			// Handle the account error.
			ai_convert.SetAIStatusQuotaExhausted(ctx)
		case "1302", "1303", "1304", "1305":
			// Handle the rate limit error.
			ai_convert.SetAIStatusExceeded(ctx)
		default:
			// Handle the bad request error.
			ai_convert.SetAIStatusInvalidRequest(ctx)
		}
	case 401:
		// Handle the auth error.
		ai_convert.SetAIStatusInvalid(ctx)
	case 400:
		// Handle the bad request error.
		ai_convert.SetAIStatusInvalidRequest(ctx)
	default:
		// Handle the bad request error.
		ai_convert.SetAIStatusInvalidRequest(ctx)

	}
}
