package google

import (
	"encoding/json"
	"strings"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	context_label "github.com/eolinker/apinto/common/context-label"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
)

//调用 Google Gemini API（以 `[https://generativelanguage.googleapis.com/v1beta](https://generativelanguage.googleapis.com/v1beta)` 为前缀）时，系统主要使用标准的 **HTTP 状态码** 结合 JSON 格式的错误响应体。
//
//响应中的错误体通常采用如下结构：
//
//```json
//{
//  "error": {
//    "code": 400,
//    "message": "API key not valid. Please pass a valid API key.",
//    "status": "INVALID_ARGUMENT",
//    "details": [...]
//  }
//}
//
//```
//
//---
//
//### 常见状态码、错误类型及常见原因
//
//| HTTP 状态码 | Error Status | 常见原因及排查方向 |
//| --- | --- | --- |
//| **400** | `INVALID_ARGUMENT` | **请求参数有误或格式不合法**：<br>
//
//<br>• JSON 格式解析失败（如多了逗号或少括号）。<br>
//
//<br>• 参数名称写错（如将 `contents` 拼错）。<br>
//
//<br>• 模型不支持所传参数或传入了越界的参数（如 `temperature` 超出 `0.0 - 2.0` 范围）。 |
//| **400** | `FAILED_PRECONDITION` | **前提条件未满足**：<br>
//
//<br>• 尝试在不支持某种功能（如 System Instruction 或 Function Calling）的旧模型上使用该功能。<br>
//
//<br>• 计费状态异常或账号尚未启用某些高级功能。 |
//| **401** | `UNAUTHENTICATED` | **身份验证失败**：<br>
//
//<br>• 缺少 API Key。<br>
//
//<br>• 传递的 API Key 无效、拼写错误或已废弃。 |
//| **403** | `PERMISSION_DENIED` | **权限不足或访问受限**：<br>
//
//<br>• 使用了被禁用的 API Key。<br>
//
//<br>• 调用的 API 或模型在当前 GCP 项目/区域未开放访问。<br>
//
//<br>• 关联的 Google Cloud 账号未开启结算（Billing）。 |
//| **404** | `NOT_FOUND` | **资源不存在**：<br>
//
//<br>• 模型 Endpoint 路径拼写错误（如 `models/gemini-1.5-pro` 写错为 `models/gemini-pro-1.5`）。<br>
//
//<br>• 调用的 Model ID 已废弃或不存在。 |
//| **429** | `RESOURCE_EXHAUSTED` | **请求触发限流（Rate Limit / Quota Exceeded）**：<br>
//
//<br>• 触发了免费层（Free Tier）的 RPM（每分钟请求数）或 TPM（每分钟 Token 数）限制。<br>
//
//<br>• 当前项目的额度用尽或短时间内并发量过高。<br>
//
//<br>• *响应头通常包含 `Retry-After`，提示需要等待的秒数。* |
//| **500** | `INTERNAL` | **服务端内部错误**：<br>
//
//<br>• Google 端模型服务出现临时故障或异常。<br>
//
//<br>• *通常只需重试（建议配合指数退避策略）。* |
//| **503** | `UNAVAILABLE` | **服务暂不可用 / 超载**：<br>
//
//<br>• 模型当前负载过高，暂时无法响应处理。<br>
//
//<br>• *建议使用带随机抖动的指数退避（Exponential Backoff）重试。* |
//| **504** | `DEADLINE_EXCEEDED` | **超时**：<br>
//
//<br>• 上游处理超时（生成极长内容或复杂推理时可能触发）。 |
//
//---
//
//### 生成内容中的特殊拦截：Safety & Block Reason
//
//有时 HTTP 状态码为 **`200 OK`**，但响应体中的 `candidates` 为空，或 `finishReason` 显示为非正常结束。这属于**安全审核机制**拦截，需要特殊处理：
//
//* **`finishReason: "SAFETY"`**：生成的回答触发了 Google 的安全策略（如敏感词、暴力、仇恨言论等）。
//* **`finishReason: "RECITATION"`**：生成的文本触发了版权或原文引用防护拦截。
//* **`finishReason: "MAX_TOKENS"`**：生成达到了设定的 `maxOutputTokens` 限制，导致内容截断（非报错，但需关注）。
//
//---
//
//### 推荐的客户端重试策略
//
//1. **处理 `429` / `500` / `503**`：实现 **带随机抖动的指数退避（Exponential Backoff with Jitter）** 进行自动重试。
//2. **处理 `400` / `401` / `403` / `404**`：直接抛出异常并告警，不需要自动重试，因为这些属于配置或请求参数错误。
//3. **解析 `candidates[0].finishReason**`：确保在 HTTP 200 返回时对非 `STOP` 状态做降级或提示处理。

type geminiErrorDetail struct {
	Type     string            `json:"@type"`
	Reason   string            `json:"reason"`
	Domain   string            `json:"domain"`
	Metadata map[string]string `json:"metadata"`
}

type geminiErrorPayload struct {
	Error struct {
		Code    int                 `json:"code"`
		Message string              `json:"message"`
		Status  string              `json:"status"`
		Details []geminiErrorDetail `json:"details"`
	} `json:"error"`
}

func ensureFailure(ctx http_context.IHttpContext) {
	// 如果已经显式触发了超时，优先保留超时状态并视为失败
	if context_label.IsAITimeout(ctx) {
		context_label.SetAIFailure(ctx, true)
		ai_convert.SetAIStatusTimeout(ctx)
		ai_convert.SetAIProviderStatuses(ctx, ai_convert.StatusTimeout)
		return
	}

	statusCode := ctx.Response().StatusCode()

	// 1. 200 OK 及正常响应：不视为失败
	if statusCode >= 200 && statusCode < 300 {
		context_label.SetAIFailure(ctx, false)
		ai_convert.SetAIStatusNormal(ctx)
		return
	}

	body := ctx.Response().GetBody()
	bodyLower := strings.ToLower(string(body))

	var errResp geminiErrorPayload
	_ = json.Unmarshal(body, &errResp)

	statusUpper := strings.ToUpper(errResp.Error.Status)
	msgLower := strings.ToLower(errResp.Error.Message)

	// 辅助检查是否包含认证/APIKey/权限相关标识
	isAuthOrPermissionError := func() bool {
		if statusUpper == "UNAUTHENTICATED" || statusUpper == "PERMISSION_DENIED" {
			return true
		}
		for _, detail := range errResp.Error.Details {
			reasonUpper := strings.ToUpper(detail.Reason)
			if strings.Contains(reasonUpper, "API_KEY") ||
				strings.Contains(reasonUpper, "AUTH") ||
				strings.Contains(reasonUpper, "PERMISSION") ||
				strings.Contains(reasonUpper, "CREDENTIAL") ||
				strings.Contains(reasonUpper, "BILLING") ||
				strings.Contains(reasonUpper, "ACCESS_TOKEN") ||
				strings.Contains(reasonUpper, "SERVICE_DISABLED") ||
				strings.Contains(reasonUpper, "CONSUMER_INVALID") {
				return true
			}
		}
		// 检查错误消息和响应体中的关键词（Google API 可能会返回带有 invalid api key 等提示的 400 错误）
		authKeywords := []string{
			"api key not valid",
			"api key expired",
			"api key is invalid",
			"invalid api key",
			"please pass a valid api key",
			"api_key_invalid",
			"apikey",
			"permission denied",
			"permission_denied",
			"unauthenticated",
			"billing",
			"access token",
			"service disabled",
			"service_disabled",
			"unregistered callers",
			"consumer identity",
			"caller does not have required permission",
			"forbidden",
		}
		for _, kw := range authKeywords {
			if strings.Contains(msgLower, kw) || strings.Contains(bodyLower, kw) {
				return true
			}
		}
		return false
	}

	// 2. 身份验证与权限失败 (401, 403, 或 400 响应中包含 API Key 无效/权限不足等信息)：视为失败
	if statusCode == 401 || statusCode == 403 || isAuthOrPermissionError() {
		context_label.SetAIFailure(ctx, true)
		ai_convert.SetAIStatusInvalid(ctx)
		ai_convert.SetAIProviderStatuses(ctx, ai_convert.StatusInvalid)
		return
	}

	// 3. 限流与配额用尽 (429 或 RESOURCE_EXHAUSTED)：视为失败
	if statusCode == 429 || statusUpper == "RESOURCE_EXHAUSTED" || strings.Contains(bodyLower, "resource_exhausted") {
		context_label.SetAIFailure(ctx, true)
		isQuota := strings.Contains(msgLower, "quota") || strings.Contains(bodyLower, "quota")
		if !isQuota {
			for _, detail := range errResp.Error.Details {
				if strings.Contains(strings.ToUpper(detail.Reason), "QUOTA") {
					isQuota = true
					break
				}
			}
		}
		if isQuota {
			ai_convert.SetAIStatusQuotaExhausted(ctx)
			ai_convert.SetAIProviderStatuses(ctx, ai_convert.StatusQuotaExhausted)
		} else {
			ai_convert.SetAIStatusExceeded(ctx)
			ai_convert.SetAIProviderStatuses(ctx, ai_convert.StatusExceeded)
		}
		return
	}

	// 4. 超时 (504 或 DEADLINE_EXCEEDED)：视为失败
	if statusCode == 504 || statusUpper == "DEADLINE_EXCEEDED" || strings.Contains(bodyLower, "deadline_exceeded") {
		context_label.SetAIFailure(ctx, true)
		ai_convert.SetAIStatusTimeout(ctx)
		ai_convert.SetAIProviderStatuses(ctx, ai_convert.StatusTimeout)
		return
	}

	// 5. 服务端错误 (5xx 或 INTERNAL / UNAVAILABLE)：视为失败
	if (statusCode >= 500 && statusCode < 600) || statusUpper == "INTERNAL" || statusUpper == "UNAVAILABLE" {
		context_label.SetAIFailure(ctx, true)
		ai_convert.SetAIStatus(ctx, "server error")
		ai_convert.SetAIProviderStatuses(ctx, "server error")
		return
	}

	// 6. 参数相关/业务相关错误 (如常规 400 INVALID_ARGUMENT、404 NOT_FOUND、旧模型不支持功能等)：不视为失败
	context_label.SetAIFailure(ctx, false)
	ai_convert.SetAIStatusInvalidRequest(ctx)
	ai_convert.SetAIProviderStatuses(ctx, ai_convert.StatusInvalidRequest)
}
