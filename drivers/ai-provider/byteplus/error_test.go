package byteplus

import (
	"fmt"
	"net/http"
	"testing"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	context_label "github.com/eolinker/apinto/common/context-label"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

type mockResponse struct {
	http_service.IResponse
	statusCode int
	body       []byte
	headers    map[string]string
}

func (m *mockResponse) StatusCode() int {
	return m.statusCode
}

func (m *mockResponse) GetBody() []byte {
	return m.body
}

func (m *mockResponse) SetBody(body []byte) {
	m.body = body
}

func (m *mockResponse) Headers() http.Header {
	h := make(http.Header)
	for k, v := range m.headers {
		h.Set(k, v)
	}
	return h
}

type mockHttpContext struct {
	http_service.IHttpContext
	response *mockResponse
	values   map[any]any
	labels   map[string]string
}

func (m *mockHttpContext) Response() http_service.IResponse {
	return m.response
}

func (m *mockHttpContext) Value(key any) any {
	if m.values == nil {
		return nil
	}
	return m.values[key]
}

func (m *mockHttpContext) WithValue(key, val any) {
	if m.values == nil {
		m.values = make(map[any]any)
	}
	m.values[key] = val
}

func (m *mockHttpContext) SetLabel(key, val string) {
	if m.labels == nil {
		m.labels = make(map[string]string)
	}
	m.labels[key] = val
}

func (m *mockHttpContext) GetLabel(key string) string {
	if m.labels == nil {
		return ""
	}
	return m.labels[key]
}

func (m *mockHttpContext) Assert(i any) error {
	if v, ok := i.(*http_service.IHttpContext); ok {
		*v = m
		return nil
	}
	return fmt.Errorf("not http context")
}

func (m *mockHttpContext) SetBalance(handler eocontext.BalanceHandler) {}

func newTestMockHttpContext(statusCode int, body string) *mockHttpContext {
	return &mockHttpContext{
		response: &mockResponse{
			statusCode: statusCode,
			body:       []byte(body),
			headers:    make(map[string]string),
		},
		values: make(map[any]any),
		labels: make(map[string]string),
	}
}

func TestEnsureFailure(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		body           string
		beforeHook     func(ctx http_service.IHttpContext)
		expectFailed   bool
		expectAIStatus string
	}{
		{
			name:           "200 OK 正常响应",
			statusCode:     200,
			body:           `{"choices":[{"message":{"role":"assistant","content":"Hello" modelark"}}]}`,
			expectFailed:   false,
			expectAIStatus: ai_convert.StatusNormal,
		},
		{
			name:       "超时标记优先判定为超时失败",
			statusCode: 200,
			body:       `{}`,
			beforeHook: func(ctx http_service.IHttpContext) {
				context_label.SetAITimeout(ctx, true)
			},
			expectFailed:   true,
			expectAIStatus: ai_convert.StatusTimeout,
		},
		{
			name:       "401 AuthenticationError 鉴权失败（OpenAI 格式）",
			statusCode: 401,
			body: `{
  "error": {
    "code": "AuthenticationError",
    "message": "API key or AK/SK is invalid.",
    "type": "Unauthorized"
  }
}`,
			expectFailed:   true,
			expectAIStatus: ai_convert.StatusInvalid,
		},
		{
			name:       "401 InvalidApiKey 密钥无效（火山引擎公共网关格式）",
			statusCode: 401,
			body: `{
  "ResponseMetadata": {
    "RequestId": "req-123",
    "Error": {
      "Code": "InvalidApiKey",
      "Message": "Invalid API-key provided."
    }
  }
}`,
			expectFailed:   true,
			expectAIStatus: ai_convert.StatusInvalid,
		},
		{
			name:       "403 AccessDenied 权限不足",
			statusCode: 403,
			body: `{
  "error": {
    "code": "AccessDenied",
    "message": "Access denied.",
    "type": "Forbidden"
  }
}`,
			expectFailed:   true,
			expectAIStatus: ai_convert.StatusInvalid,
		},
		{
			name:       "403 OperationDenied.ServiceNotOpen 服务未开通",
			statusCode: 403,
			body: `{
  "error": {
    "code": "OperationDenied.ServiceNotOpen",
    "message": "The service is not opened, please activate the model service in console.",
    "type": "Forbidden"
  }
}`,
			expectFailed:   true,
			expectAIStatus: ai_convert.StatusInvalid,
		},
		{
			name:       "403 OperationDenied.ServiceOverdue 余额欠费",
			statusCode: 403,
			body: `{
  "error": {
    "code": "OperationDenied.ServiceOverdue",
    "message": "Your account balance is overdue.",
    "type": "Forbidden"
  }
}`,
			expectFailed:   true,
			expectAIStatus: ai_convert.StatusInvalid,
		},
		{
			name:       "400 Arrearage 账户欠费访问被拒（官方 400 状态码但属于账号欠费）",
			statusCode: 400,
			body: `{
  "error": {
    "code": "Arrearage",
    "message": "Access denied, please make sure your account is in good standing."
  }
}`,
			expectFailed:   true,
			expectAIStatus: ai_convert.StatusInvalid,
		},
		{
			name:       "400 InvalidEndpoint.ClosedEndpoint 推理端点已关闭（端点基础设施故障）",
			statusCode: 400,
			body: `{
  "error": {
    "code": "InvalidEndpoint.ClosedEndpoint",
    "message": "The endpoint is closed or temporarily unavailable.",
    "type": "BadRequest"
  }
}`,
			expectFailed:   true,
			expectAIStatus: ai_convert.StatusInvalid,
		},
		{
			name:       "400 InvalidSubscription 订阅未开通或已过期",
			statusCode: 400,
			body: `{
  "error": {
    "code": "InvalidSubscription",
    "message": "Coding plan subscription is invalid or expired.",
    "type": "Forbidden"
  }
}`,
			expectFailed:   true,
			expectAIStatus: ai_convert.StatusInvalid,
		},
		{
			name:       "400 MissingParameter 缺少请求参数（常规业务错误，不视为失败）",
			statusCode: 400,
			body: `{
  "error": {
    "code": "MissingParameter",
    "message": "Required parameter 'model' is missing.",
    "type": "BadRequest"
  }
}`,
			expectFailed:   false,
			expectAIStatus: ai_convert.StatusInvalidRequest,
		},
		{
			name:       "400 InvalidParameter 参数错误（常规业务错误，不视为失败）",
			statusCode: 400,
			body: `{
  "error": {
    "code": "InvalidParameter",
    "message": "One or more parameters specified in the request are not valid.",
    "type": "BadRequest",
    "param": "temperature"
  }
}`,
			expectFailed:   false,
			expectAIStatus: ai_convert.StatusInvalidRequest,
		},
		{
			name:       "400 SensitiveContentDetected 内容安全拦截（业务风控，不视为基础设施失败）",
			statusCode: 400,
			body: `{
  "error": {
    "code": "SensitiveContentDetected",
    "message": "The input content violates the safety policy.",
    "type": "BadRequest"
  }
}`,
			expectFailed:   false,
			expectAIStatus: ai_convert.StatusInvalidRequest,
		},
		{
			name:       "400 DataInspectionFailed 风控规则拦截（不视为基础设施失败）",
			statusCode: 400,
			body: `{
  "error": {
    "code": "DataInspectionFailed",
    "message": "Input or output data may contain inappropriate content.",
    "type": "BadRequest"
  }
}`,
			expectFailed:   false,
			expectAIStatus: ai_convert.StatusInvalidRequest,
		},
		{
			name:       "400 OutofContextError 上下文超长（业务错误，不视为失败）",
			statusCode: 400,
			body: `{
  "error": {
    "code": "OutofContextError",
    "message": "Total tokens exceed the context window limit.",
    "type": "BadRequest"
  }
}`,
			expectFailed:   false,
			expectAIStatus: ai_convert.StatusInvalidRequest,
		},
		{
			name:       "404 InvalidEndpointOrModel.NotFound 模型或端点不存在（业务输入错误，不视为失败）",
			statusCode: 404,
			body: `{
  "error": {
    "code": "InvalidEndpointOrModel.NotFound",
    "message": "The specified model ep-xxx could not be found.",
    "type": "NotFound"
  }
}`,
			expectFailed:   false,
			expectAIStatus: ai_convert.StatusInvalidRequest,
		},
		{
			name:       "429 RateLimitExceeded.EndpointRPMExceeded 接入点速率限制（视为失败）",
			statusCode: 429,
			body: `{
  "error": {
    "code": "RateLimitExceeded.EndpointRPMExceeded",
    "message": "The request has exceeded the RPM limit for endpoint.",
    "type": "TooManyRequests"
  }
}`,
			expectFailed:   true,
			expectAIStatus: ai_convert.StatusExceeded,
		},
		{
			name:       "429 Throttling 限流（视为失败）",
			statusCode: 429,
			body: `{
  "error": {
    "code": "Throttling",
    "message": "Requests throttling triggered."
  }
}`,
			expectFailed:   true,
			expectAIStatus: ai_convert.StatusExceeded,
		},
		{
			name:       "429 QuotaExceeded 额度用尽（视为失败）",
			statusCode: 429,
			body: `{
  "error": {
    "code": "QuotaExceeded",
    "message": "Free allocated quota exceeded, please purchase or increase quota.",
    "type": "TooManyRequests"
  }
}`,
			expectFailed:   true,
			expectAIStatus: ai_convert.StatusQuotaExhausted,
		},
		{
			name:       "429 Throttling.AllocationQuota 配额耗尽（视为失败）",
			statusCode: 429,
			body: `{
  "error": {
    "code": "Throttling.AllocationQuota",
    "message": "Allocated quota exceeded."
  }
}`,
			expectFailed:   true,
			expectAIStatus: ai_convert.StatusQuotaExhausted,
		},
		{
			name:       "408 RequestTimeOut 请求超时（视为失败）",
			statusCode: 408,
			body: `{
  "error": {
    "code": "RequestTimeOut",
    "message": "Request timed out, please try again later."
  }
}`,
			expectFailed:   true,
			expectAIStatus: ai_convert.StatusTimeout,
		},
		{
			name:       "504 DeadlineExceeded 上游网关超时（视为失败）",
			statusCode: 504,
			body: `{
  "error": {
    "code": "DeadlineExceeded",
    "message": "Gateway timeout."
  }
}`,
			expectFailed:   true,
			expectAIStatus: ai_convert.StatusTimeout,
		},
		{
			name:       "500 InternalError 服务端内部错误（视为失败）",
			statusCode: 500,
			body: `{
  "error": {
    "code": "InternalError",
    "message": "An internal error has occurred, please try again later.",
    "type": "InternalError"
  }
}`,
			expectFailed:   true,
			expectAIStatus: "server error",
		},
		{
			name:       "503 ModelUnavailable 模型不可用或过载（视为失败）",
			statusCode: 503,
			body: `{
  "error": {
    "code": "ModelUnavailable",
    "message": "The model service is temporarily unavailable or overloaded.",
    "type": "Unavailable"
  }
}`,
			expectFailed:   true,
			expectAIStatus: "server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newTestMockHttpContext(tt.statusCode, tt.body)
			if tt.beforeHook != nil {
				tt.beforeHook(ctx)
			}

			ensureFailure(ctx)

			actualFailed := context_label.IsAIFailure(ctx)
			if actualFailed != tt.expectFailed {
				t.Errorf("[%s] expected isAIFailure=%v, got=%v", tt.name, tt.expectFailed, actualFailed)
			}

			actualStatus := ai_convert.GetAIStatus(ctx)
			if actualStatus != tt.expectAIStatus {
				t.Errorf("[%s] expected aiStatus=%s, got=%s", tt.name, tt.expectAIStatus, actualStatus)
			}
		})
	}
}

func TestErrorCallback(t *testing.T) {
	ctx := newTestMockHttpContext(401, `{"error":{"code":"InvalidApiKey","message":"api key invalid"}}`)
	errorCallback(ctx, []byte(`{"error":{"code":"InvalidApiKey","message":"api key invalid"}}`))

	if !context_label.IsAIFailure(ctx) {
		t.Errorf("expected isAIFailure=true")
	}
	if ai_convert.GetAIStatus(ctx) != ai_convert.StatusInvalid {
		t.Errorf("expected status %s, got %s", ai_convert.StatusInvalid, ai_convert.GetAIStatus(ctx))
	}
}
