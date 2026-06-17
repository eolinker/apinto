package variable_relation

import (
	"net"
	"strings"
	"testing"

	http_context_node "github.com/eolinker/apinto/node/http-context"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/valyala/fasthttp"
)

// mockCustomerVar mock 实现 eosc.ICustomerVar
type mockCustomerVar struct {
	data map[string]map[string]string
}

func (m *mockCustomerVar) Exists(key string, field string) bool {
	if fields, ok := m.data[key]; ok {
		_, okField := fields[field]
		return okField
	}
	return false
}

func (m *mockCustomerVar) Get(key string, field string) (string, bool) {
	if fields, ok := m.data[key]; ok {
		val, okField := fields[field]
		return val, okField
	}
	return "", false
}

func (m *mockCustomerVar) GetAll(key string) (map[string]string, bool) {
	if fields, ok := m.data[key]; ok {
		return fields, true
	}
	return nil, false
}

// initTestContext 初始化测试用的 IHttpContext
func initTestContext(method, path string, labels map[string]string) (http_service.IHttpContext, error) {
	fast := &fasthttp.RequestCtx{}
	freq := fasthttp.AcquireRequest()
	freq.Header.SetMethod(method)
	freq.URI().SetPath(path)

	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
	if err != nil {
		return nil, err
	}
	fast.Init(freq, addr, nil)

	ctx := http_context_node.NewContext(fast, 0)
	for k, v := range labels {
		ctx.SetLabel(k, v)
	}
	return ctx, nil
}

func TestParseTemplate(t *testing.T) {
	cases := []struct {
		tpl  string
		want []Segment
	}{
		{
			tpl: "{method} {path}",
			want: []Segment{
				{IsVar: true, Name: "method"},
				{IsVar: false, Value: " "},
				{IsVar: true, Name: "path"},
			},
		},
		{
			tpl: "{app_id}:{path}",
			want: []Segment{
				{IsVar: true, Name: "app_id"},
				{IsVar: false, Value: ":"},
				{IsVar: true, Name: "path"},
			},
		},
		{
			tpl: "{path}",
			want: []Segment{
				{IsVar: true, Name: "path"},
			},
		},
	}

	for _, cc := range cases {
		got := ParseTemplate(cc.tpl)
		if len(got) != len(cc.want) {
			t.Fatalf("tpl %s: got len %d, want %d", cc.tpl, len(got), len(cc.want))
		}
		for i := range got {
			if got[i].IsVar != cc.want[i].IsVar || got[i].Name != cc.want[i].Name || got[i].Value != cc.want[i].Value {
				t.Errorf("tpl %s segment %d: got %+v, want %+v", cc.tpl, i, got[i], cc.want[i])
			}
		}
	}
}

func TestExtractValues(t *testing.T) {
	cases := []struct {
		name    string
		tpl     string
		pattern string
		want    map[string]string
		ok      bool
	}{
		{
			name:    "method and path",
			tpl:     "{method} {path}",
			pattern: "GET /api/v1/user",
			want: map[string]string{
				"method": "GET",
				"path":   "/api/v1/user",
			},
			ok: true,
		},
		{
			name:    "app_id and path",
			tpl:     "{app_id}:{path}",
			pattern: "user_service:/api/user/{id}",
			want: map[string]string{
				"app_id": "user_service",
				"path":   "/api/user/{id}",
			},
			ok: true,
		},
		{
			name:    "single path",
			tpl:     "{path}",
			pattern: "/api/user",
			want: map[string]string{
				"path": "/api/user",
			},
			ok: true,
		},
		{
			name:    "mismatch static separator",
			tpl:     "{app_id}:{path}",
			pattern: "user_service /api/user",
			want:    nil,
			ok:      false,
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			segs := ParseTemplate(cc.tpl)
			got, ok := ExtractValues(cc.pattern, segs)
			if ok != cc.ok {
				t.Fatalf("ExtractValues ok %t, want %t", ok, cc.ok)
			}
			if ok {
				for k, v := range cc.want {
					if got[k] != v {
						t.Errorf("key %s: got %s, want %s", k, got[k], v)
					}
				}
			}
		})
	}
}

// mockChain mock Chain
type mockChain struct {
	called bool
}

func (m *mockChain) DoChain(ctx eocontext.EoContext) error {
	m.called = true
	return nil
}

func (m *mockChain) Destroy() {}

func TestDoHttpFilter(t *testing.T) {
	// mock customerVar
	mockVar := &mockCustomerVar{
		data: map[string]map[string]string{
			"service_bind_api:user_service": {
				"GET /api/v1/user":  "api_get_user",
				"POST /api/v1/user": "api_create_user",
			},
			"app_bind_api:client_app": {
				"GET /api/v2/task/{id}": "api_get_task",
			},
		},
	}
	customerVar = mockVar

	factory := NewFactory()

	cases := []struct {
		name          string
		method        string
		path          string
		labels        map[string]string
		rules         []*Rule
		wantApiID     string
		wantTaskApiID string
		wantErrMsg    string
	}{
		{
			name:   "match service rule successfully",
			method: "GET",
			path:   "/api/v1/user",
			labels: map[string]string{
				"service": "user_service",
			},
			rules: []*Rule{
				{
					Key:        "service_bind_api:{service}",
					InnerKey:   "{method} {path}",
					ValueLabel: "api_id",
				},
			},
			wantApiID:  "api_get_user",
			wantErrMsg: "",
		},
		{
			name:   "match client app rule successfully",
			method: "GET",
			path:   "/api/v2/task/123",
			labels: map[string]string{
				"app_id": "client_app",
			},
			rules: []*Rule{
				{
					Key:        "app_bind_api:{app_id}",
					InnerKey:   "{method} {path}",
					ValueLabel: "task_api_id",
				},
			},
			wantTaskApiID: "api_get_task",
			wantErrMsg:    "",
		},
		{
			name:   "mapping not found error",
			method: "GET",
			path:   "/api/v1/user",
			labels: map[string]string{
				"service": "non_existent_service",
			},
			rules: []*Rule{
				{
					Key:        "service_bind_api:{service}",
					InnerKey:   "{method} {path}",
					ValueLabel: "api_id",
				},
			},
			wantErrMsg: "no mapping found for key",
		},
		{
			name:   "no pattern matched error",
			method: "DELETE",
			path:   "/api/v1/user",
			labels: map[string]string{
				"service": "user_service",
			},
			rules: []*Rule{
				{
					Key:        "service_bind_api:{service}",
					InnerKey:   "{method} {path}",
					ValueLabel: "api_id",
				},
			},
			wantErrMsg: "no pattern matched for key",
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			config := &Config{
				Rules: cc.rules,
			}

			driver, err := factory.Create("variable_relation", "variable_relation", "variable_relation", "desc", nil)
			if err != nil {
				t.Fatalf("factory create driver error: %v", err)
			}

			worker, err := driver.Create("plugin@instance", "variable_relation", config, nil)
			if err != nil {
				t.Fatalf("create worker error: %v", err)
			}

			filter, ok := worker.(http_service.HttpFilter)
			if !ok {
				t.Fatalf("worker is not http_service.HttpFilter")
			}

			ctx, err := initTestContext(cc.method, cc.path, cc.labels)
			if err != nil {
				t.Fatalf("init context error: %v", err)
			}

			chain := &mockChain{}
			err = filter.DoHttpFilter(ctx, chain)

			if cc.wantErrMsg != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, but got nil", cc.wantErrMsg)
				}
				if !strings.Contains(err.Error(), cc.wantErrMsg) {
					t.Errorf("expected error containing %q, but got %q", cc.wantErrMsg, err.Error())
				}
				if chain.called {
					t.Errorf("chain should not be called on error")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !chain.called {
					t.Errorf("chain should be called successfully")
				}
				if cc.wantApiID != "" {
					got := ctx.GetLabel("api_id")
					if got != cc.wantApiID {
						t.Errorf("want api_id %s, got %s", cc.wantApiID, got)
					}
				}
				if cc.wantTaskApiID != "" {
					got := ctx.GetLabel("task_api_id")
					if got != cc.wantTaskApiID {
						t.Errorf("want task_api_id %s, got %s", cc.wantTaskApiID, got)
					}
				}
			}
		})
	}
}
