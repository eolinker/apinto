package google

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ai_convert "github.com/eolinker/apinto/ai-convert"
	context_label "github.com/eolinker/apinto/common/context-label"
	"golang.org/x/oauth2"
)

func generateTestServiceAccountBase64(t *testing.T, tokenURI string) string {
	t.Helper()
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate rsa key: %v", err)
	}
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		t.Fatalf("failed to marshal pkcs8 private key: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	})

	sa := map[string]string{
		"type":         "service_account",
		"project_id":   "xroute-473401",
		"private_key":  string(pemBytes),
		"client_email": "test@xroute-473401.iam.gserviceaccount.com",
		"token_uri":    tokenURI,
	}
	raw, err := json.Marshal(sa)
	if err != nil {
		t.Fatalf("failed to marshal service account json: %v", err)
	}
	return base64.StdEncoding.EncodeToString(raw)
}

func TestNewVideoTaskCommit(t *testing.T) {
	tests := []struct {
		name         string
		baseUrl      string
		expectedPath string
		expectErr    bool
	}{
		{
			name:         "默认 BaseUrl 为空",
			baseUrl:      "",
			expectedPath: "/v1beta/interactions",
			expectErr:    false,
		},
		{
			name:         "仅包含域名的 BaseUrl",
			baseUrl:      "https://aiplatform.googleapis.com",
			expectedPath: "/v1beta1/interactions",
			expectErr:    false,
		},
		{
			name:         "带 project_id 与 locations 的 BaseUrl",
			baseUrl:      "https://aiplatform.googleapis.com/v1beta1/projects/xroute-473401/locations/global",
			expectedPath: "/v1beta1/projects/xroute-473401/locations/global/interactions",
			expectErr:    false,
		},
		{
			name:         "BaseUrl 末尾带斜杠",
			baseUrl:      "https://aiplatform.googleapis.com/v1beta1/projects/xroute-473401/locations/global/",
			expectedPath: "/v1beta1/projects/xroute-473401/locations/global/interactions",
			expectErr:    false,
		},
		{
			name:         "BaseUrl 已包含 /interactions 后缀",
			baseUrl:      "https://aiplatform.googleapis.com/v1beta1/projects/xroute-473401/locations/global/interactions",
			expectedPath: "/v1beta1/projects/xroute-473401/locations/global/interactions",
			expectErr:    false,
		},
		{
			name:      "无效 BaseUrl",
			baseUrl:   "::invalid-url::",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver, err := NewVideoTaskCommit(provider, "test-sa-key", tt.baseUrl, ai_convert.ModelTypeVideoGeneration, 10*time.Minute)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			v, ok := driver.(*Video)
			if !ok {
				t.Fatalf("expected *Video, got %T", driver)
			}
			if v.Provider() != provider {
				t.Errorf("expected provider %s, got %s", provider, v.Provider())
			}
			if v.ModelType() != ai_convert.ModelTypeVideoGeneration {
				t.Errorf("expected modelType %s, got %s", ai_convert.ModelTypeVideoGeneration, v.ModelType())
			}
			if v.path != tt.expectedPath {
				t.Errorf("expected path %s, got %s", tt.expectedPath, v.path)
			}
		})
	}
}

func TestVideoRequestConvert_WithMockOAuthServer(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"mock-oauth2-token-xyz","token_type":"Bearer","expires_in":3600}`))
	}))
	defer tokenServer.Close()

	saKeyBase64 := generateTestServiceAccountBase64(t, tokenServer.URL)
	baseUrl := "https://aiplatform.googleapis.com/v1beta1/projects/xroute-473401/locations/global"

	driver, err := NewVideoTaskCommit(provider, saKeyBase64, baseUrl, ai_convert.ModelTypeVideoGeneration, 10*time.Minute)
	if err != nil {
		t.Fatalf("NewVideoTaskCommit failed: %v", err)
	}

	reqBody := []byte(`{"model":"gemini-omni-1.1-flash-preview","input":"Generate a video of a cat"}`)
	mCtx := &mockHttpContext{
		proxy: &mockRequest{
			header: &mockHeader{},
			body:   &mockBody{body: reqBody},
			uri:    &mockURI{},
		},
		values:    make(map[interface{}]interface{}),
		labels:    make(map[string]string),
		requestId: "req-video-1",
	}

	err = driver.RequestConvert(mCtx, nil)
	if err != nil {
		t.Fatalf("RequestConvert failed: %v", err)
	}

	authHeader := mCtx.proxy.header.Get("Authorization")
	if authHeader != "Bearer mock-oauth2-token-xyz" {
		t.Errorf("expected Authorization header 'Bearer mock-oauth2-token-xyz', got '%s'", authHeader)
	}

	expectedPath := "/v1beta1/projects/xroute-473401/locations/global/interactions"
	if mCtx.proxy.uri.path != expectedPath {
		t.Errorf("expected URI path '%s', got '%s'", expectedPath, mCtx.proxy.uri.path)
	}

	if mCtx.balanceHandler == nil {
		t.Errorf("expected balanceHandler to be set on context")
	}

	if context_label.GetBillingMode(mCtx) != context_label.BillingModeImmediate {
		t.Errorf("expected billing mode %s, got %s", context_label.BillingModeImmediate, context_label.GetBillingMode(mCtx))
	}
}

func TestVideoRequestConvert_ReuseValidToken(t *testing.T) {
	driver, err := NewVideoTaskCommit(provider, "unused-sa-key", "https://aiplatform.googleapis.com/v1beta1/projects/xroute-473401/locations/global", ai_convert.ModelTypeVideoGeneration, 10*time.Minute)
	if err != nil {
		t.Fatalf("NewVideoTaskCommit failed: %v", err)
	}

	v := driver.(*Video)
	v.token = &oauth2.Token{
		AccessToken: "cached-access-token",
		Expiry:      time.Now().Add(1 * time.Hour),
	}

	mCtx := &mockHttpContext{
		proxy: &mockRequest{
			header: &mockHeader{},
			body:   &mockBody{},
			uri:    &mockURI{},
		},
		values: make(map[interface{}]interface{}),
		labels: make(map[string]string),
	}

	if err := v.RequestConvert(mCtx, nil); err != nil {
		t.Fatalf("RequestConvert failed: %v", err)
	}

	if got := mCtx.proxy.header.Get("Authorization"); got != "Bearer cached-access-token" {
		t.Errorf("expected cached token 'Bearer cached-access-token', got '%s'", got)
	}
}

func TestVideoRequestConvert_ErrorCases(t *testing.T) {
	t.Run("无效的 Base64 saKey", func(t *testing.T) {
		driver, err := NewVideoTaskCommit(provider, "!!!not-base64!!!", "", ai_convert.ModelTypeVideoGeneration, 10*time.Minute)
		if err != nil {
			t.Fatalf("NewVideoTaskCommit failed: %v", err)
		}
		mCtx := &mockHttpContext{}
		if err := driver.RequestConvert(mCtx, nil); err == nil {
			t.Errorf("expected error for invalid base64 saKey, got nil")
		}
	})

	t.Run("无效的 ServiceAccount JSON", func(t *testing.T) {
		invalidSA := base64.StdEncoding.EncodeToString([]byte(`{"invalid":"json"}`))
		driver, err := NewVideoTaskCommit(provider, invalidSA, "", ai_convert.ModelTypeVideoGeneration, 10*time.Minute)
		if err != nil {
			t.Fatalf("NewVideoTaskCommit failed: %v", err)
		}
		mCtx := &mockHttpContext{}
		if err := driver.RequestConvert(mCtx, nil); err == nil {
			t.Errorf("expected error for invalid service account json, got nil")
		}
	})
}

func TestVideoResponseConvert(t *testing.T) {
	driver, err := NewVideoTaskCommit(provider, "sa-key", "", ai_convert.ModelTypeVideoGeneration, 10*time.Minute)
	if err != nil {
		t.Fatalf("NewVideoTaskCommit failed: %v", err)
	}

	t.Run("200 成功响应", func(t *testing.T) {
		ctx := newTestMockHttpContext(200, `{"id":"inter-123","status":"completed"}`)
		if err := driver.ResponseConvert(ctx); err != nil {
			t.Fatalf("ResponseConvert failed: %v", err)
		}
		if context_label.IsAIFailure(ctx) {
			t.Errorf("expected IsAIFailure=false on 200")
		}
		if ai_convert.GetAIStatus(ctx) != ai_convert.StatusNormal {
			t.Errorf("expected status %s, got %s", ai_convert.StatusNormal, ai_convert.GetAIStatus(ctx))
		}
	})

	t.Run("401 鉴权失败响应", func(t *testing.T) {
		ctx := newTestMockHttpContext(401, `{"error":{"code":401,"message":"Request had invalid authentication credentials.","status":"UNAUTHENTICATED"}}`)
		if err := driver.ResponseConvert(ctx); err != nil {
			t.Fatalf("ResponseConvert failed: %v", err)
		}
		if !context_label.IsAIFailure(ctx) {
			t.Errorf("expected IsAIFailure=true on 401")
		}
		if ai_convert.GetAIStatus(ctx) != ai_convert.StatusInvalid {
			t.Errorf("expected status %s, got %s", ai_convert.StatusInvalid, ai_convert.GetAIStatus(ctx))
		}
	})
}

func TestFactoryCreate_VideoGeneration(t *testing.T) {
	cfgJSON := `{"api_key":"test-api-key","service_account_key":"dGVzdC1zYS1rZXk=","base_url":"https://aiplatform.googleapis.com/v1beta1/projects/xroute-473401/locations/global"}`
	converter, err := Create(cfgJSON)
	if err != nil {
		t.Fatalf("Create converter failed: %v", err)
	}

	videoDriver, ok := converter.Get(ai_convert.ModelTypeVideoGeneration)
	if !ok || videoDriver == nil {
		t.Fatalf("expected driver for ModelTypeVideoGeneration to be registered")
	}
	if videoDriver.ModelType() != ai_convert.ModelTypeVideoGeneration {
		t.Errorf("expected ModelType %s, got %s", ai_convert.ModelTypeVideoGeneration, videoDriver.ModelType())
	}
}
