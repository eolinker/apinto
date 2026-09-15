package byteplus

import (
	"encoding/json"
	"fmt"
	"strings"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	context_label "github.com/eolinker/apinto/common/context-label"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

// 参考 BytePlus ModelArk 官方错误码文档：
// https://docs.byteplus.com/en/docs/modelark/1299023
//
// 调用 BytePlus ModelArk API 时，系统使用标准 HTTP 状态码结合 JSON 格式的错误响应体。
// 错误体通常兼容 OpenAI 格式或火山引擎/BytePlus 原生网关结构：
//
// 格式一（OpenAI 兼容）：
// ```json
// {
//   "error": {
//     "code": "InvalidParameter",
//     "message": "One or more parameters specified in the request are not valid. Request ID: ...",
//     "type": "BadRequest",
//     "param": "..."
//   }
// }
// ```
//
// 格式二（火山引擎 / BytePlus 公共网关）：
// ```json
// {
//   "ResponseMetadata": {
//     "RequestId": "...",
//     "Action": "...",
//     "Version": "...",
//     "Service": "...",
//     "Region": "...",
//     "Error": {
//       "CodeN": 100000,
//       "Code": "InvalidParameter",
//       "Message": "The request contains invalid parameters."
//     }
//   }
// }
// ```
//
// ---
//
// ### 常见状态码、错误类型及排查方向
//
// | HTTP 状态码 | Error Type | Error Code | 常见原因及排查方向 |
// | --- | --- | --- | --- |
// | **400** | `BadRequest` | `MissingParameter` | 请求缺少必填参数。 |
// | **400** | `BadRequest` | `InvalidParameter` | 请求参数不合法，或包含不支持的配置。 |
// | **400** | `BadRequest` | `InvalidEndpoint.ClosedEndpoint` | 推理接入点已关闭或暂时不可用（端点故障）。 |
// | **400** | `BadRequest` | `SensitiveContentDetected` / `*.PolicyViolation` / `*.PrivacyInformation` / `*.DeepFake` | 输入或输出包含敏感词、违规版权或隐私信息（内容安全风控拦截，属业务请求拦截，不视为基础设施失败）。 |
// | **400** | `BadRequest` | `InvalidArgumentError` / `UnknownRole` / `InvalidImageDetail` / `InvalidPixelLimit` | 消息缺少 role、角色未定义、图片详情或像素参数不合法。 |
// | **400** | `BadRequest` | `InvalidImageURL.EmptyURL` / `InvalidImageURL.InvalidFormat` | Base64 图片 URL 为空或格式不受支持。 |
// | **400** | `BadRequest` | `OutofContextError` | 文本与图片总 token 超出模型上下文上限。 |
// | **400** | `Forbidden` | `InvalidSubscription` | 账号未开通 Coding Plan 订阅或已过期。 |
// | **400** | - | `Arrearage` | 账户欠费，服务受限。 |
// | **401** | `Unauthorized` | `AuthenticationError` | API Key 或 AK/SK 缺失或无效。 |
// | **401** | `Forbidden` | `InvalidAccountStatus` | 账号状态异常。 |
// | **401** | - | `InvalidApiKey` | 传递的 API Key 无效。 |
// | **403** | `Forbidden` | `AccessDenied` / `OperationDenied.PermissionDenied` | 权限不足，无法访问指定模型或资源。 |
// | **403** | `Forbidden` | `OperationDenied.ServiceNotOpen` | 模型服务未开通，需在 ModelArk 控制台开通服务。 |
// | **403** | `Forbidden` | `OperationDenied.ServiceOverdue` / `AccountOverdueError` | 账户余额欠费逾期，需前往充值中心充值。 |
// | **403** | `Forbidden` | `OperationDenied.FileQuotaExceeded` | 文件存储配额已耗尽。 |
// | **403** | `Forbidden` | `OperationDenied.ArkAccessRoleNotFound` / `TosAccessDenied` | TOS 资源或项目授权角色未配置。 |
// | **404** | `NotFound` | `InvalidEndpointOrModel.NotFound` / `NotFound.*` | 模型或接入点不存在（业务输入错误）。 |
// | **404** | `NotFound` | `ModelNotOpen` | 账号未开通该模型服务（权限/开通问题）。 |
// | **404** | `NotFound` | `UnsupportedModel` | 模型不支持当前调用功能（业务请求错误）。 |
// | **408** | - | `RequestTimeOut` | 请求超时。 |
// | **413** | `BadRequest` | `Payload Too Large` | 上传载荷超出限制。 |
// | **415** | `BadRequest` | `BadRequest.UnsupportedFileFormat` | 输入文件格式不受支持。 |
// | **429** | `TooManyRequests` | `RateLimitExceeded.EndpointRPMExceeded` / `EndpointTPMExceeded` | 接入点 RPM / TPM 速率达到限制。 |
// | **429** | `TooManyRequests` | `ModelAccountRpmRateLimitExceeded` / `ModelAccountTpmRateLimitExceeded` / `APIAccountRpmRateLimitExceeded` / `ModelAccountIpmRateLimitExceeded` | 账号/模型级别 RPM、TPM、IPM 限流。 |
// | **429** | `TooManyRequests` | `QuotaExceeded` / `Throttling.AllocationQuota` | 免费额度或配额用尽。 |
// | **500** | `InternalError` | `InternalError` / `SystemError` / `InternalServiceError` | 服务端内部错误。 |
// | **500** | `InternalError` | `InternalError.Timeout` | 服务端内部超时。 |
// | **503** | `Unavailable` | `ModelUnavailable` / `ServerOverloaded` | 模型服务暂不可用或过载。 |
// | **504** | `GatewayTimeout` | `DeadlineExceeded` / `DEADLINE_EXCEEDED` | 上游超时。 |

type byteplusErrorPayload struct {
	Error struct {
		Code    any    `json:"code"`
		Message string `json:"message"`
		Type    string `json:"type"`
		Param   string `json:"param"`
	} `json:"error"`
	ResponseMetadata struct {
		RequestId string `json:"RequestId"`
		Action    string `json:"Action"`
		Version   string `json:"Version"`
		Service   string `json:"Service"`
		Region    string `json:"Region"`
		Error     struct {
			CodeN   int    `json:"CodeN"`
			Code    string `json:"Code"`
			Message string `json:"Message"`
		} `json:"Error"`
	} `json:"ResponseMetadata"`
	Code    any    `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

func ensureFailure(ctx http_service.IHttpContext) {
	// 0. 如果已经显式触发了超时，优先保留超时状态并视为失败
	if context_label.IsAITimeout(ctx) {
		context_label.SetAIFailure(ctx, true)
		ai_convert.SetAIStatusTimeout(ctx)
		ai_convert.SetAIProviderStatuses(ctx, ai_convert.StatusTimeout)
		return
	}

	statusCode := ctx.Response().StatusCode()

	// 1. 200 OK 及 2xx 正常响应：不视为失败
	if statusCode >= 200 && statusCode < 300 {
		context_label.SetAIFailure(ctx, false)
		ai_convert.SetAIStatusNormal(ctx)
		return
	}

	body := ctx.Response().GetBody()
	bodyLower := strings.ToLower(string(body))

	var errResp byteplusErrorPayload
	_ = json.Unmarshal(body, &errResp)

	// 统一提取 Code、Message、Type
	codeStr := ""
	if errResp.Error.Code != nil {
		codeStr = fmt.Sprintf("%v", errResp.Error.Code)
	} else if errResp.ResponseMetadata.Error.Code != "" {
		codeStr = errResp.ResponseMetadata.Error.Code
	} else if errResp.Code != nil {
		codeStr = fmt.Sprintf("%v", errResp.Code)
	}

	msgStr := errResp.Error.Message
	if msgStr == "" {
		msgStr = errResp.ResponseMetadata.Error.Message
	}
	if msgStr == "" {
		msgStr = errResp.Message
	}

	typeStr := errResp.Error.Type
	if typeStr == "" {
		typeStr = errResp.Type
	}

	codeUpper := strings.ToUpper(codeStr)
	msgLower := strings.ToLower(msgStr)
	typeUpper := strings.ToUpper(typeStr)

	// 辅助判断是否属于认证、密钥、权限、服务未开通、欠费或订阅过期等基础设施错误
	isAuthOrPermissionError := func() bool {
		if statusCode == 401 || statusCode == 403 {
			return true
		}
		if typeUpper == "UNAUTHORIZED" || typeUpper == "FORBIDDEN" {
			return true
		}
		authCodeKeywords := []string{
			"AUTHENTICATIONERROR",
			"INVALIDAPIKEY",
			"INVALIDACCOUNTSTATUS",
			"ACCESSDENIED",
			"PERMISSIONDENIED",
			"SERVICENOTOPEN",
			"MODELNOTOPEN",
			"SERVICEOVERDUE",
			"ACCOUNTOVERDUEERROR",
			"ARREARAGE",
			"INVALIDSUBSCRIPTION",
			"UNPURCHASED",
			"COMMODITYNOTPURCHASED",
			"ARKACCESSROLENOTFOUND",
			"TOSACCESSDENIED",
		}
		for _, kw := range authCodeKeywords {
			if strings.Contains(codeUpper, kw) {
				return true
			}
		}
		authMsgKeywords := []string{
			"api key", "apikey", "ak/sk", "authentication",
			"unauthorized", "access denied", "permission denied",
			"account status", "not activated the model", "not open",
			"service is unavailable, please go to the",
			"balance is overdue", "overdue balance", "recharge",
			"subscription", "arrearage",
		}
		for _, kw := range authMsgKeywords {
			if strings.Contains(msgLower, kw) || strings.Contains(bodyLower, kw) {
				return true
			}
		}
		return false
	}

	// 2. 身份验证、密钥、权限、服务开通状态、欠费及订阅类错误：视为失败
	if isAuthOrPermissionError() {
		context_label.SetAIFailure(ctx, true)
		ai_convert.SetAIStatusInvalid(ctx)
		ai_convert.SetAIProviderStatuses(ctx, ai_convert.StatusInvalid)
		return
	}

	// 3. 特殊端点故障（如 HTTP 400 下 InvalidEndpoint.ClosedEndpoint 表示推理接入点已关闭或临时不可用，属端点故障）：视为失败
	if strings.Contains(codeUpper, "CLOSEDENDPOINT") || strings.Contains(msgLower, "closed or temporarily unavailable") {
		context_label.SetAIFailure(ctx, true)
		ai_convert.SetAIStatusInvalid(ctx)
		ai_convert.SetAIProviderStatuses(ctx, ai_convert.StatusInvalid)
		return
	}

	// 4. 限流与配额用尽 (429 或含有 RateLimit/Quota/Throttling)：视为失败
	is429 := statusCode == 429 || typeUpper == "TOOMANYREQUESTS" ||
		strings.Contains(codeUpper, "RATELIMIT") ||
		strings.Contains(codeUpper, "QUOTA") ||
		strings.Contains(codeUpper, "THROTTLING") ||
		strings.Contains(codeUpper, "LIMITREQUESTS")

	if is429 {
		context_label.SetAIFailure(ctx, true)

		isQuota := strings.Contains(codeUpper, "QUOTA") ||
			strings.Contains(codeUpper, "ALLOCATIONQUOTA") ||
			strings.Contains(codeUpper, "FILEQUOTAEXCEEDED") ||
			strings.Contains(codeUpper, "BILLOVERDUE") ||
			strings.Contains(msgLower, "quota") ||
			strings.Contains(msgLower, "exhausted") ||
			strings.Contains(msgLower, "free trial") ||
			strings.Contains(bodyLower, "quota")

		if isQuota {
			ai_convert.SetAIStatusQuotaExhausted(ctx)
			ai_convert.SetAIProviderStatuses(ctx, ai_convert.StatusQuotaExhausted)
		} else {
			ai_convert.SetAIStatusExceeded(ctx)
			ai_convert.SetAIProviderStatuses(ctx, ai_convert.StatusExceeded)
		}
		return
	}

	// 5. 超时 (408 / 504 / RequestTimeOut / DeadlineExceeded / InternalError.Timeout)：视为失败
	if statusCode == 408 || statusCode == 504 ||
		strings.Contains(codeUpper, "TIMEOUT") ||
		strings.Contains(codeUpper, "DEADLINE_EXCEEDED") ||
		strings.Contains(bodyLower, "deadline_exceeded") {
		context_label.SetAIFailure(ctx, true)
		ai_convert.SetAIStatusTimeout(ctx)
		ai_convert.SetAIProviderStatuses(ctx, ai_convert.StatusTimeout)
		return
	}

	// 6. 服务端内部错误与过载 (5xx 或 InternalError / SystemError / ModelUnavailable / ServerOverloaded)：视为失败
	if (statusCode >= 500 && statusCode < 600) ||
		strings.Contains(codeUpper, "INTERNALERROR") ||
		strings.Contains(codeUpper, "SYSTEMERROR") ||
		strings.Contains(codeUpper, "INTERNALSERVICEERROR") ||
		strings.Contains(codeUpper, "MODELUNAVAILABLE") ||
		strings.Contains(codeUpper, "SERVEROVERLOADED") ||
		strings.Contains(codeUpper, "MODELSERVICEFAILED") ||
		strings.Contains(codeUpper, "INVOKEPLUGINFAILED") ||
		strings.Contains(codeUpper, "APPPROCESSFAILED") {
		context_label.SetAIFailure(ctx, true)
		ai_convert.SetAIStatus(ctx, "server error")
		ai_convert.SetAIProviderStatuses(ctx, "server error")
		return
	}

	// 7. 参数相关与业务输入错误（不视为失败）：
	// 包含常规 400 参数缺失/不合法、OutofContextError、404 模型或资源不存在、413、415，
	// 以及内容安全审查拦截（SensitiveContentDetected、PolicyViolation、PrivacyInformation、DeepFake 等）。
	context_label.SetAIFailure(ctx, false)
	ai_convert.SetAIStatusInvalidRequest(ctx)
	ai_convert.SetAIProviderStatuses(ctx, ai_convert.StatusInvalidRequest)
}

func errorCallback(ctx http_service.IHttpContext, body []byte) {
	ensureFailure(ctx)
}
