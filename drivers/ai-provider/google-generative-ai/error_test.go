package google

import (
	"testing"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	context_label "github.com/eolinker/apinto/common/context-label"
)

func newTestMockHttpContext(statusCode int, body string) *mockHttpContext {
	ctx := &mockHttpContext{
		response: &mockResponse{
			statusCode: statusCode,
			body:       []byte(body),
		},
		values: make(map[interface{}]interface{}),
		labels: make(map[string]string),
	}
	return ctx
}

func TestEnsureFailure(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		body           string
		expectFailed   bool
		expectAIStatus string
	}{
		{
			name:           "200 OK 正常响应",
			statusCode:     200,
			body:           `{"candidates":[{"content":{"parts":[{"text":"Hello!"}]}}]}`,
			expectFailed:   false,
			expectAIStatus: ai_convert.StatusNormal,
		},
		{
			name:       "400 INVALID_ARGUMENT 但带有 API key not valid（注释中典型示例，需判定为失败）",
			statusCode: 400,
			body: `{
  "error": {
    "code": 400,
    "message": "API key not valid. Please pass a valid API key.",
    "status": "INVALID_ARGUMENT",
    "details": [
      {
        "@type": "type.googleapis.com/google.rpc.ErrorInfo",
        "reason": "API_KEY_INVALID",
        "domain": "googleapis.com"
      }
    ]
  }
}`,
			expectFailed:   true,
			expectAIStatus: ai_convert.StatusInvalid,
		},
		{
			name:       "400 INVALID_ARGUMENT 常规参数错误（contents 字段拼错，不视为失败）",
			statusCode: 400,
			body: `{
  "error": {
    "code": 400,
    "message": "Invalid JSON payload received. Unknown name \"contents_err\": Cannot find field.",
    "status": "INVALID_ARGUMENT"
  }
}`,
			expectFailed:   false,
			expectAIStatus: ai_convert.StatusInvalidRequest,
		},
		{
			name:       "400 INVALID_ARGUMENT 参数越界（如 temperature 超出范围，不视为失败）",
			statusCode: 400,
			body: `{
  "error": {
    "code": 400,
    "message": "temperature must be between 0.0 and 2.0",
    "status": "INVALID_ARGUMENT"
  }
}`,
			expectFailed:   false,
			expectAIStatus: ai_convert.StatusInvalidRequest,
		},
		{
			name:       "400 FAILED_PRECONDITION 旧模型不支持某特性（业务错误，不视为失败）",
			statusCode: 400,
			body: `{
  "error": {
    "code": 400,
    "message": "Model does not support system instructions.",
    "status": "FAILED_PRECONDITION"
  }
}`,
			expectFailed:   false,
			expectAIStatus: ai_convert.StatusInvalidRequest,
		},
		{
			name:       "400 FAILED_PRECONDITION 未开启结算计费（权限/账号问题，视为失败）",
			statusCode: 400,
			body: `{
  "error": {
    "code": 400,
    "message": "User location is not supported for the API use without billing.",
    "status": "FAILED_PRECONDITION"
  }
}`,
			expectFailed:   true,
			expectAIStatus: ai_convert.StatusInvalid,
		},
		{
			name:       "401 UNAUTHENTICATED 身份验证失败（API Key 缺失或无效，视为失败）",
			statusCode: 401,
			body: `{
  "error": {
    "code": 401,
    "message": "Request had invalid authentication credentials.",
    "status": "UNAUTHENTICATED"
  }
}`,
			expectFailed:   true,
			expectAIStatus: ai_convert.StatusInvalid,
		},
		{
			name:       "403 PERMISSION_DENIED 权限不足或被禁用（视为失败）",
			statusCode: 403,
			body: `{
  "error": {
    "code": 403,
    "message": "The caller does not have permission",
    "status": "PERMISSION_DENIED"
  }
}`,
			expectFailed:   true,
			expectAIStatus: ai_convert.StatusInvalid,
		},
		{
			name:       "404 NOT_FOUND 模型不存在或路径拼写错误（业务参数错误，不视为失败）",
			statusCode: 404,
			body: `{
  "error": {
    "code": 404,
    "message": "models/gemini-pro-1.5 is not found for API version v1beta",
    "status": "NOT_FOUND"
  }
}`,
			expectFailed:   false,
			expectAIStatus: ai_convert.StatusInvalidRequest,
		},
		{
			name:       "429 RESOURCE_EXHAUSTED 配额不足 quota（视为失败）",
			statusCode: 429,
			body: `{
  "error": {
    "code": 429,
    "message": "Resource has been exhausted (e.g. check quota).",
    "status": "RESOURCE_EXHAUSTED"
  }
}`,
			expectFailed:   true,
			expectAIStatus: ai_convert.StatusQuotaExhausted,
		},
		{
			name:       "429 RESOURCE_EXHAUSTED 速率超限 rate limit（视为失败）",
			statusCode: 429,
			body: `{
  "error": {
    "code": 429,
    "message": "Rate limit exceeded for requests per minute.",
    "status": "RESOURCE_EXHAUSTED"
  }
}`,
			expectFailed:   true,
			expectAIStatus: ai_convert.StatusExceeded,
		},
		{
			name:       "500 INTERNAL 服务端内部错误（视为失败）",
			statusCode: 500,
			body: `{
  "error": {
    "code": 500,
    "message": "An internal error has occurred.",
    "status": "INTERNAL"
  }
}`,
			expectFailed:   true,
			expectAIStatus: "server error",
		},
		{
			name:       "503 UNAVAILABLE 服务超载暂不可用（视为失败）",
			statusCode: 503,
			body: `{
  "error": {
    "code": 503,
    "message": "The service is temporarily unavailable.",
    "status": "UNAVAILABLE"
  }
}`,
			expectFailed:   true,
			expectAIStatus: "server error",
		},
		{
			name:       "504 DEADLINE_EXCEEDED 超时（视为失败）",
			statusCode: 504,
			body: `{
  "error": {
    "code": 504,
    "message": "Deadline exceeded.",
    "status": "DEADLINE_EXCEEDED"
  }
}`,
			expectFailed:   true,
			expectAIStatus: ai_convert.StatusTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newTestMockHttpContext(tt.statusCode, tt.body)
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
