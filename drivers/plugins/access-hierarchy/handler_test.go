package access_hierarchy

import (
	"strconv"
	"testing"
	"time"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/eocontext"
	http_context "github.com/eolinker/eosc/eocontext/http-context"
	"github.com/eolinker/eosc/metrics"
	testifyAssert "github.com/stretchr/testify/assert"
)

type mockMetrics struct {
	val string
}

func (m *mockMetrics) Metrics(entry metrics.LabelReader) string {
	return m.val
}

func (m *mockMetrics) Key() string {
	return m.val
}

type mockEntry struct{}

func (m *mockEntry) Children(child string) []eosc.IEntry {
	return nil
}

func (m *mockEntry) Read(pattern string) interface{} {
	return ""
}

func (m *mockEntry) ReadLabel(pattern string) string {
	return ""
}

type mockCustomerVar struct {
	data map[string]map[string]string
}

func (m *mockCustomerVar) Exists(key string, field string) bool {
	fields, has := m.data[key]
	if !has {
		return false
	}
	_, ok := fields[field]
	return ok
}

func (m *mockCustomerVar) Get(key string, field string) (string, bool) {
	fields, has := m.data[key]
	if !has {
		return "", false
	}
	val, ok := fields[field]
	return val, ok
}

func (m *mockCustomerVar) GetAll(key string) (map[string]string, bool) {
	fields, has := m.data[key]
	return fields, has
}

func TestHandler_Check(t *testing.T) {
	oldVar := customerVar
	defer func() { customerVar = oldVar }()

	t.Run("empty app or resource ID", func(t *testing.T) {
		mockVar := &mockCustomerVar{data: make(map[string]map[string]string)}
		customerVar = mockVar

		h := &handler{
			a: &mockMetrics{val: ""},
			b: &mockMetrics{val: "res1"},
		}
		testifyAssert.False(t, h.Check(&mockEntry{}))

		h2 := &handler{
			a: &mockMetrics{val: "app1"},
			b: &mockMetrics{val: ""},
		}
		testifyAssert.False(t, h2.Check(&mockEntry{}))
	})

	t.Run("no app groups bound", func(t *testing.T) {
		mockVar := &mockCustomerVar{data: make(map[string]map[string]string)}
		customerVar = mockVar

		h := &handler{
			a: &mockMetrics{val: "app1"},
			b: &mockMetrics{val: "res1"},
		}
		testifyAssert.False(t, h.Check(&mockEntry{}))
	})

	t.Run("simple explicit resource match pass", func(t *testing.T) {
		mockVar := &mockCustomerVar{
			data: map[string]map[string]string{
				"app_groups:app1": {
					"group1": "0",
				},
				"group_meta:group1": {
					"mode": "explicit_resources",
				},
				"group_resources:group1": {
					"res1": "0",
				},
			},
		}
		customerVar = mockVar

		h := &handler{
			a: &mockMetrics{val: "app1"},
			b: &mockMetrics{val: "res1"},
		}
		testifyAssert.True(t, h.Check(&mockEntry{}))
	})

	t.Run("simple explicit resource expired app-group", func(t *testing.T) {
		pastTime := time.Now().UnixMilli() - 10000
		mockVar := &mockCustomerVar{
			data: map[string]map[string]string{
				"app_groups:app1": {
					"group1": strconv.FormatInt(pastTime, 10),
				},
				"group_meta:group1": {
					"mode": "explicit_resources",
				},
				"group_resources:group1": {
					"res1": "0",
				},
			},
		}
		customerVar = mockVar

		h := &handler{
			a: &mockMetrics{val: "app1"},
			b: &mockMetrics{val: "res1"},
		}
		testifyAssert.False(t, h.Check(&mockEntry{}))
	})

	t.Run("simple explicit resource expired group-resource", func(t *testing.T) {
		pastTime := time.Now().UnixMilli() - 10000
		mockVar := &mockCustomerVar{
			data: map[string]map[string]string{
				"app_groups:app1": {
					"group1": "0",
				},
				"group_meta:group1": {
					"mode": "explicit_resources",
				},
				"group_resources:group1": {
					"res1": strconv.FormatInt(pastTime, 10),
				},
			},
		}
		customerVar = mockVar

		h := &handler{
			a: &mockMetrics{val: "app1"},
			b: &mockMetrics{val: "res1"},
		}
		testifyAssert.False(t, h.Check(&mockEntry{}))
	})

	t.Run("follow tenant mode pass", func(t *testing.T) {
		mockVar := &mockCustomerVar{
			data: map[string]map[string]string{
				"app_groups:app1": {
					"group1": "0",
				},
				"group_meta:group1": {
					"mode":            "follow_tenant",
					"owner_tenant_id": "tenant1",
				},
				"tenant_groups:tenant1": {
					"group1": "0",
					"group2": "0",
				},
				"group_meta:group2": {
					"mode": "explicit_resources",
				},
				"group_resources:group2": {
					"res1": "0",
				},
			},
		}
		customerVar = mockVar

		h := &handler{
			a: &mockMetrics{val: "app1"},
			b: &mockMetrics{val: "res1"},
		}
		testifyAssert.True(t, h.Check(&mockEntry{}))
	})

	t.Run("follow tenant mode filter non-explicit owned group", func(t *testing.T) {
		mockVar := &mockCustomerVar{
			data: map[string]map[string]string{
				"app_groups:app1": {
					"group1": "0",
				},
				"group_meta:group1": {
					"mode":            "follow_tenant",
					"owner_tenant_id": "tenant1",
				},
				"tenant_groups:tenant1": {
					"group2": "0",
				},
				"group_meta:group2": {
					"mode": "follow_tenant",
				},
				"group_resources:group2": {
					"res1": "0",
				},
			},
		}
		customerVar = mockVar

		h := &handler{
			a: &mockMetrics{val: "app1"},
			b: &mockMetrics{val: "res1"},
		}
		testifyAssert.False(t, h.Check(&mockEntry{}))
	})

	t.Run("multi-level hierarchical granted groups pass", func(t *testing.T) {
		mockVar := &mockCustomerVar{
			data: map[string]map[string]string{
				"app_groups:app1": {
					"group1": "0",
				},
				"group_meta:group1": {
					"mode":            "follow_tenant",
					"owner_tenant_id": "tenant_child",
				},
				"tenant_granted_groups:tenant_child": {
					"group_parent": "0",
				},
				"group_meta:group_parent": {
					"mode":            "follow_tenant",
					"owner_tenant_id": "tenant_parent",
				},
				"tenant_groups:tenant_parent": {
					"group_grandparent": "0",
				},
				"group_meta:group_grandparent": {
					"mode": "explicit_resources",
				},
				"group_resources:group_grandparent": {
					"res1": "0",
				},
			},
		}
		customerVar = mockVar

		h := &handler{
			a: &mockMetrics{val: "app1"},
			b: &mockMetrics{val: "res1"},
		}
		testifyAssert.True(t, h.Check(&mockEntry{}))
	})

	t.Run("cycle detection group cycle", func(t *testing.T) {
		mockVar := &mockCustomerVar{
			data: map[string]map[string]string{
				"app_groups:app1": {
					"group1": "0",
				},
				"group_meta:group1": {
					"mode":            "follow_tenant",
					"owner_tenant_id": "tenant1",
				},
				"tenant_granted_groups:tenant1": {
					"group2": "0",
				},
				"group_meta:group2": {
					"mode":            "follow_tenant",
					"owner_tenant_id": "tenant2",
				},
				"tenant_granted_groups:tenant2": {
					"group1": "0",
				},
			},
		}
		customerVar = mockVar

		h := &handler{
			a: &mockMetrics{val: "app1"},
			b: &mockMetrics{val: "res_nonexistent"},
		}
		testifyAssert.False(t, h.Check(&mockEntry{}))
	})

	t.Run("cycle detection tenant cycle", func(t *testing.T) {
		mockVar := &mockCustomerVar{
			data: map[string]map[string]string{
				"app_groups:app1": {
					"group1": "0",
				},
				"group_meta:group1": {
					"mode":            "follow_tenant",
					"owner_tenant_id": "tenant1",
				},
				"tenant_granted_groups:tenant1": {
					"group2": "0",
				},
				"group_meta:group2": {
					"mode":            "follow_tenant",
					"owner_tenant_id": "tenant1",
				},
			},
		}
		customerVar = mockVar

		h := &handler{
			a: &mockMetrics{val: "app1"},
			b: &mockMetrics{val: "res_nonexistent"},
		}
		testifyAssert.False(t, h.Check(&mockEntry{}))
	})
}

func TestAccessHierarchy_WorkerMethods(t *testing.T) {
	ah := &AccessHierarchy{}
	testifyAssert.NoError(t, ah.Start())
	testifyAssert.NoError(t, ah.Stop())
	ah.Destroy()

	testifyAssert.True(t, ah.CheckSkill(http_context.FilterSkillName))
	testifyAssert.False(t, ah.CheckSkill("some_other_skill"))
}

func TestAccessHierarchy_AssertAndParse(t *testing.T) {
	ah := &AccessHierarchy{}

	cfg := &Config{
		Rules: []*Rule{
			{
				A: "app",
				B: "res",
			},
		},
	}

	res, err := assert(cfg)
	testifyAssert.NoError(t, err)
	testifyAssert.Equal(t, cfg, res)

	res2, err2 := assert("invalid_type")
	testifyAssert.Error(t, err2)
	testifyAssert.Nil(t, res2)

	resp, rules := ah.parseConfig(cfg)
	testifyAssert.NotNil(t, resp)
	testifyAssert.Len(t, rules, 1)
}

type mockExtenderDriverRegister struct {
	name    string
	factory eosc.IExtenderDriverFactory
}

func (m *mockExtenderDriverRegister) RegisterExtenderDriver(name string, factory eosc.IExtenderDriverFactory) error {
	m.name = name
	m.factory = factory
	return nil
}

type mockHttpResponse struct {
	http_context.IResponse
	status  int
	headers map[string]string
	body    []byte
}

func (m *mockHttpResponse) SetStatus(code int, text string) {
	m.status = code
}

func (m *mockHttpResponse) SetHeader(key string, value string) {
	if m.headers == nil {
		m.headers = make(map[string]string)
	}
	m.headers[key] = value
}

func (m *mockHttpResponse) SetBody(body []byte) {
	m.body = body
}

type mockHttpContext struct {
	http_context.IHttpContext
	labels map[string]string
	values map[string]interface{}
	resp   *mockHttpResponse
}

func (m *mockHttpContext) GetLabel(name string) string {
	return m.labels[name]
}

func (m *mockHttpContext) Value(name interface{}) interface{} {
	if s, ok := name.(string); ok {
		return m.values[s]
	}
	return nil
}

func (m *mockHttpContext) Response() http_context.IResponse {
	return m.resp
}

func (m *mockHttpContext) Assert(i interface{}) error {
	if ref, ok := i.(*http_context.IHttpContext); ok {
		*ref = m
		return nil
	}
	return nil
}

type mockChain struct {
	called bool
}

func (m *mockChain) DoChain(ctx eocontext.EoContext) error {
	m.called = true
	return nil
}

func (m *mockChain) Destroy() {}

func TestFactoryAndRegister(t *testing.T) {
	reg := &mockExtenderDriverRegister{}
	Register(reg)
	testifyAssert.Equal(t, Name, reg.name)
	testifyAssert.NotNil(t, reg.factory)

	fact := NewFactory()
	testifyAssert.NotNil(t, fact)

	err := Check(&Config{}, nil)
	testifyAssert.NoError(t, err)

	worker, err := Create("id1", "name1", &Config{
		Rules: []*Rule{
			{A: "app", B: "res"},
		},
	}, nil)
	testifyAssert.NoError(t, err)
	testifyAssert.NotNil(t, worker)
}

func TestAccessHierarchy_Reset(t *testing.T) {
	ah := &AccessHierarchy{}
	err := ah.Reset(&Config{
		Rules: []*Rule{
			{A: "app", B: "res"},
		},
	}, nil)
	testifyAssert.NoError(t, err)
	testifyAssert.Len(t, ah.rules, 1)

	err = ah.Reset("invalid", nil)
	testifyAssert.Error(t, err)
}

func TestAccessHierarchy_Filter(t *testing.T) {
	ah := &AccessHierarchy{
		response: defaultResponse,
	}

	chain := &mockChain{}
	ctx := &mockHttpContext{
		labels: map[string]string{},
		values: map[string]interface{}{},
		resp:   &mockHttpResponse{},
	}
	err := ah.DoHttpFilter(ctx, chain)
	testifyAssert.NoError(t, err)
	testifyAssert.True(t, chain.called)

	oldVar := customerVar
	defer func() { customerVar = oldVar }()
	customerVar = &mockCustomerVar{
		data: map[string]map[string]string{
			"app_groups:app1": {
				"group1": "0",
			},
			"group_meta:group1": {
				"mode": "explicit_resources",
			},
			"group_resources:group1": {
				"res1": "0",
			},
		},
	}

	ah.rules = []ruleHandler{
		ah.newHandler("$ctx_app", "$ctx_res", nil),
	}

	chain2 := &mockChain{}
	ctx2 := &mockHttpContext{
		labels: map[string]string{
			"app": "app1",
			"res": "res1",
		},
		resp: &mockHttpResponse{},
	}

	err = ah.DoHttpFilter(ctx2, chain2)
	testifyAssert.NoError(t, err)
	testifyAssert.True(t, chain2.called)

	chain3 := &mockChain{}
	ctx3 := &mockHttpContext{
		labels: map[string]string{
			"app": "app1",
			"res": "res_nonexistent",
		},
		resp: &mockHttpResponse{},
	}
	err = ah.DoHttpFilter(ctx3, chain3)
	testifyAssert.NoError(t, err)
	testifyAssert.False(t, chain3.called)
	testifyAssert.Equal(t, 403, ctx3.resp.status)

	chain4 := &mockChain{}
	ctx4 := &mockHttpContext{
		labels: map[string]string{
			"app": "app1",
			"res": "res1",
		},
		resp: &mockHttpResponse{},
	}
	err = ah.DoFilter(ctx4, chain4)
	testifyAssert.NoError(t, err)
	testifyAssert.True(t, chain4.called)
}

func TestHandler_EdgeCasesAdditional(t *testing.T) {
	oldVar := customerVar
	defer func() { customerVar = oldVar }()

	t.Run("group_meta not found", func(t *testing.T) {
		mockVar := &mockCustomerVar{
			data: map[string]map[string]string{
				"app_groups:app1": {
					"group1": "0",
				},
			},
		}
		customerVar = mockVar

		h := &handler{
			a: &mockMetrics{val: "app1"},
			b: &mockMetrics{val: "res1"},
		}
		testifyAssert.False(t, h.Check(&mockEntry{}))
	})

	t.Run("group_resources not found for explicit_resources", func(t *testing.T) {
		mockVar := &mockCustomerVar{
			data: map[string]map[string]string{
				"app_groups:app1": {
					"group1": "0",
				},
				"group_meta:group1": {
					"mode": "explicit_resources",
				},
			},
		}
		customerVar = mockVar

		h := &handler{
			a: &mockMetrics{val: "app1"},
			b: &mockMetrics{val: "res1"},
		}
		testifyAssert.False(t, h.Check(&mockEntry{}))
	})

	t.Run("explicit_resources resource exists but expired", func(t *testing.T) {
		pastTime := time.Now().UnixMilli() - 10000
		mockVar := &mockCustomerVar{
			data: map[string]map[string]string{
				"app_groups:app1": {
					"group1": "0",
				},
				"group_meta:group1": {
					"mode": "explicit_resources",
				},
				"group_resources:group1": {
					"res1": strconv.FormatInt(pastTime, 10),
				},
			},
		}
		customerVar = mockVar

		h := &handler{
			a: &mockMetrics{val: "app1"},
			b: &mockMetrics{val: "res1"},
		}
		testifyAssert.False(t, h.Check(&mockEntry{}))
	})

	t.Run("explicit_resources resource not found", func(t *testing.T) {
		mockVar := &mockCustomerVar{
			data: map[string]map[string]string{
				"app_groups:app1": {
					"group1": "0",
				},
				"group_meta:group1": {
					"mode": "explicit_resources",
				},
				"group_resources:group1": {
					"other_res": "0",
				},
			},
		}
		customerVar = mockVar

		h := &handler{
			a: &mockMetrics{val: "app1"},
			b: &mockMetrics{val: "res1"},
		}
		testifyAssert.False(t, h.Check(&mockEntry{}))
	})

	t.Run("unsupported group mode", func(t *testing.T) {
		mockVar := &mockCustomerVar{
			data: map[string]map[string]string{
				"app_groups:app1": {
					"group1": "0",
				},
				"group_meta:group1": {
					"mode": "unsupported_mode",
				},
			},
		}
		customerVar = mockVar

		h := &handler{
			a: &mockMetrics{val: "app1"},
			b: &mockMetrics{val: "res1"},
		}
		testifyAssert.False(t, h.Check(&mockEntry{}))
	})

	t.Run("tenant_groups owned group expired", func(t *testing.T) {
		pastTime := time.Now().UnixMilli() - 10000
		mockVar := &mockCustomerVar{
			data: map[string]map[string]string{
				"app_groups:app1": {
					"group1": "0",
				},
				"group_meta:group1": {
					"mode":            "follow_tenant",
					"owner_tenant_id": "tenant1",
				},
				"tenant_groups:tenant1": {
					"group2": strconv.FormatInt(pastTime, 10),
				},
			},
		}
		customerVar = mockVar

		h := &handler{
			a: &mockMetrics{val: "app1"},
			b: &mockMetrics{val: "res1"},
		}
		testifyAssert.False(t, h.Check(&mockEntry{}))
	})

	t.Run("tenant_groups owned group meta not found", func(t *testing.T) {
		mockVar := &mockCustomerVar{
			data: map[string]map[string]string{
				"app_groups:app1": {
					"group1": "0",
				},
				"group_meta:group1": {
					"mode":            "follow_tenant",
					"owner_tenant_id": "tenant1",
				},
				"tenant_groups:tenant1": {
					"group2": "0",
				},
			},
		}
		customerVar = mockVar

		h := &handler{
			a: &mockMetrics{val: "app1"},
			b: &mockMetrics{val: "res1"},
		}
		testifyAssert.False(t, h.Check(&mockEntry{}))
	})

	t.Run("tenant_groups owned group resources not found", func(t *testing.T) {
		mockVar := &mockCustomerVar{
			data: map[string]map[string]string{
				"app_groups:app1": {
					"group1": "0",
				},
				"group_meta:group1": {
					"mode":            "follow_tenant",
					"owner_tenant_id": "tenant1",
				},
				"tenant_groups:tenant1": {
					"group2": "0",
				},
				"group_meta:group2": {
					"mode": "explicit_resources",
				},
			},
		}
		customerVar = mockVar

		h := &handler{
			a: &mockMetrics{val: "app1"},
			b: &mockMetrics{val: "res1"},
		}
		testifyAssert.False(t, h.Check(&mockEntry{}))
	})

	t.Run("tenant_groups owned group resources exist but expired", func(t *testing.T) {
		pastTime := time.Now().UnixMilli() - 10000
		mockVar := &mockCustomerVar{
			data: map[string]map[string]string{
				"app_groups:app1": {
					"group1": "0",
				},
				"group_meta:group1": {
					"mode":            "follow_tenant",
					"owner_tenant_id": "tenant1",
				},
				"tenant_groups:tenant1": {
					"group2": "0",
				},
				"group_meta:group2": {
					"mode": "explicit_resources",
				},
				"group_resources:group2": {
					"res1": strconv.FormatInt(pastTime, 10),
				},
			},
		}
		customerVar = mockVar

		h := &handler{
			a: &mockMetrics{val: "app1"},
			b: &mockMetrics{val: "res1"},
		}
		testifyAssert.False(t, h.Check(&mockEntry{}))
	})

	t.Run("tenant_groups owned group resources other nonexistent", func(t *testing.T) {
		mockVar := &mockCustomerVar{
			data: map[string]map[string]string{
				"app_groups:app1": {
					"group1": "0",
				},
				"group_meta:group1": {
					"mode":            "follow_tenant",
					"owner_tenant_id": "tenant1",
				},
				"tenant_groups:tenant1": {
					"group2": "0",
				},
				"group_meta:group2": {
					"mode": "explicit_resources",
				},
				"group_resources:group2": {
					"other_res": "0",
				},
			},
		}
		customerVar = mockVar

		h := &handler{
			a: &mockMetrics{val: "app1"},
			b: &mockMetrics{val: "res1"},
		}
		testifyAssert.False(t, h.Check(&mockEntry{}))
	})

	t.Run("tenant_granted_groups expired", func(t *testing.T) {
		pastTime := time.Now().UnixMilli() - 10000
		mockVar := &mockCustomerVar{
			data: map[string]map[string]string{
				"app_groups:app1": {
					"group1": "0",
				},
				"group_meta:group1": {
					"mode":            "follow_tenant",
					"owner_tenant_id": "tenant1",
				},
				"tenant_granted_groups:tenant1": {
					"group2": strconv.FormatInt(pastTime, 10),
				},
			},
		}
		customerVar = mockVar

		h := &handler{
			a: &mockMetrics{val: "app1"},
			b: &mockMetrics{val: "res1"},
		}
		testifyAssert.False(t, h.Check(&mockEntry{}))
	})
}

func TestAccessHierarchy_newHandlerEdge(t *testing.T) {
	ah := &AccessHierarchy{}
	h := ah.newHandler("ctx_app", "ctx_res", nil)
	testifyAssert.NotNil(t, h)
}

func TestAccessHierarchy_CustomConfig(t *testing.T) {
	ah := &AccessHierarchy{}
	h := ah.newHandler("$ctx_app", "$ctx_res", &HierarchyConfig{
		AppNodesPrefix:           "custom_app_groups:",
		NodeMetaPrefix:           "custom_group_meta:",
		NodeTargetsPrefix:        "custom_group_resources:",
		ParentNodesPrefix:        "custom_tenant_groups:",
		ParentGrantedNodesPrefix: "custom_tenant_granted_groups:",
		MetaModeField:            "custom_mode",
		MetaParentNodeField:      "custom_owner_tenant_id",
		ModeExplicit:             "custom_explicit",
		ModeFollowParent:         "custom_follow",
	})
	testifyAssert.NotNil(t, h)

	handlerObj := h.(*handler)
	testifyAssert.Equal(t, "custom_app_groups:", handlerObj.getAppNodesPrefix())
	testifyAssert.Equal(t, "custom_group_meta:", handlerObj.getNodeMetaPrefix())
	testifyAssert.Equal(t, "custom_group_resources:", handlerObj.getNodeTargetsPrefix())
	testifyAssert.Equal(t, "custom_tenant_groups:", handlerObj.getParentNodesPrefix())
	testifyAssert.Equal(t, "custom_tenant_granted_groups:", handlerObj.getParentGrantedNodesPrefix())
	testifyAssert.Equal(t, "custom_mode", handlerObj.getMetaModeField())
	testifyAssert.Equal(t, "custom_owner_tenant_id", handlerObj.getMetaParentNodeField())
	testifyAssert.Equal(t, "custom_explicit", handlerObj.getModeExplicit())
	testifyAssert.Equal(t, "custom_follow", handlerObj.getModeFollowParent())

	hEmpty := ah.newHandler("$ctx_app", "$ctx_res", &HierarchyConfig{})
	testifyAssert.NotNil(t, hEmpty)
	handlerEmpty := hEmpty.(*handler)
	testifyAssert.Equal(t, "app_groups:", handlerEmpty.getAppNodesPrefix())
	testifyAssert.Equal(t, "group_meta:", handlerEmpty.getNodeMetaPrefix())
	testifyAssert.Equal(t, "group_resources:", handlerEmpty.getNodeTargetsPrefix())
	testifyAssert.Equal(t, "tenant_groups:", handlerEmpty.getParentNodesPrefix())
	testifyAssert.Equal(t, "tenant_granted_groups:", handlerEmpty.getParentGrantedNodesPrefix())
	testifyAssert.Equal(t, "mode", handlerEmpty.getMetaModeField())
	testifyAssert.Equal(t, "owner_tenant_id", handlerEmpty.getMetaParentNodeField())
	testifyAssert.Equal(t, "explicit_resources", handlerEmpty.getModeExplicit())
	testifyAssert.Equal(t, "follow_tenant", handlerEmpty.getModeFollowParent())
}

func TestAccessHierarchy_CustomConfigExecution(t *testing.T) {
	oldVar := customerVar
	defer func() { customerVar = oldVar }()

	customerVar = &mockCustomerVar{
		data: map[string]map[string]string{
			"custom_app_groups:app1": {
				"group1": "0",
			},
			"custom_group_meta:group1": {
				"custom_mode": "custom_explicit",
			},
			"custom_group_resources:group1": {
				"res1": "0",
			},
		},
	}

	ah := &AccessHierarchy{}
	h := ah.newHandler("$ctx_app", "$ctx_res", &HierarchyConfig{
		AppNodesPrefix:    "custom_app_groups:",
		NodeMetaPrefix:    "custom_group_meta:",
		NodeTargetsPrefix: "custom_group_resources:",
		MetaModeField:     "custom_mode",
		ModeExplicit:      "custom_explicit",
	})

	chain := &mockChain{}
	ctx := &mockHttpContext{
		labels: map[string]string{
			"app": "app1",
			"res": "res1",
		},
		resp: &mockHttpResponse{},
	}

	ah.rules = []ruleHandler{h}
	err := ah.DoHttpFilter(ctx, chain)
	testifyAssert.NoError(t, err)
	testifyAssert.True(t, chain.called)
}

func TestAccessHierarchy_InfiniteDepthSelfInheritance(t *testing.T) {
	oldVar := customerVar
	defer func() { customerVar = oldVar }()

	mockVar := &mockCustomerVar{
		data: map[string]map[string]string{
			"user_roles:user1": {
				"role1": "0",
			},
			"role_meta:role1": {
				"mode":           "follow",
				"parent_role_id": "role2",
			},
			"role_meta:role2": {
				"mode":           "follow",
				"parent_role_id": "role3",
			},
			"role_meta:role3": {
				"mode": "explicit",
			},
			"role_permissions:role3": {
				"permission1": "0",
			},
		},
	}
	customerVar = mockVar

	ah := &AccessHierarchy{}
	h := ah.newHandler("$ctx_app", "$ctx_res", &HierarchyConfig{
		AppNodesPrefix:      "user_roles:",
		NodeMetaPrefix:      "role_meta:",
		NodeTargetsPrefix:   "role_permissions:",
		MetaModeField:       "mode",
		MetaParentNodeField: "parent_role_id",
		ModeExplicit:        "explicit",
		ModeFollowParent:    "follow",
	})

	handlerObj := h.(*handler)
	testifyAssert.True(t, handlerObj.checkNodeHasTarget("role1", "permission1", time.Now().UnixMilli(), make(map[string]bool), make(map[string]bool)))
	testifyAssert.False(t, handlerObj.checkNodeHasTarget("role1", "permission_nonexistent", time.Now().UnixMilli(), make(map[string]bool), make(map[string]bool)))
}

func TestAccessHierarchy_InfiniteDepthSelfInheritanceCycle(t *testing.T) {
	oldVar := customerVar
	defer func() { customerVar = oldVar }()

	mockVar := &mockCustomerVar{
		data: map[string]map[string]string{
			"user_roles:user1": {
				"role1": "0",
			},
			"role_meta:role1": {
				"mode":           "follow",
				"parent_role_id": "role2",
			},
			"role_meta:role2": {
				"mode":           "follow",
				"parent_role_id": "role1",
			},
		},
	}
	customerVar = mockVar

	ah := &AccessHierarchy{}
	h := ah.newHandler("$ctx_app", "$ctx_res", &HierarchyConfig{
		AppNodesPrefix:      "user_roles:",
		NodeMetaPrefix:      "role_meta:",
		NodeTargetsPrefix:   "role_permissions:",
		MetaModeField:       "mode",
		MetaParentNodeField: "parent_role_id",
		ModeExplicit:        "explicit",
		ModeFollowParent:    "follow",
	})

	handlerObj := h.(*handler)
	testifyAssert.False(t, handlerObj.checkNodeHasTarget("role1", "permission1", time.Now().UnixMilli(), make(map[string]bool), make(map[string]bool)))
}

func TestAccessHierarchy_BackwardCompatibleConfig(t *testing.T) {
	ah := &AccessHierarchy{}
	h := ah.newHandler("$ctx_app", "$ctx_res", &HierarchyConfig{
		AppGroupsPrefix:           "old_app_groups:",
		GroupMetaPrefix:           "old_group_meta:",
		GroupResourcesPrefix:      "old_group_resources:",
		TenantGroupsPrefix:        "old_tenant_groups:",
		TenantGrantedGroupsPrefix: "old_tenant_granted_groups:",
		MetaOwnerTenantField:      "old_owner_tenant_id",
		ModeFollowTenant:          "old_follow_tenant",
	})
	testifyAssert.NotNil(t, h)

	handlerObj := h.(*handler)
	testifyAssert.Equal(t, "old_app_groups:", handlerObj.getAppNodesPrefix())
	testifyAssert.Equal(t, "old_group_meta:", handlerObj.getNodeMetaPrefix())
	testifyAssert.Equal(t, "old_group_resources:", handlerObj.getNodeTargetsPrefix())
	testifyAssert.Equal(t, "old_tenant_groups:", handlerObj.getParentNodesPrefix())
	testifyAssert.Equal(t, "old_tenant_granted_groups:", handlerObj.getParentGrantedNodesPrefix())
	testifyAssert.Equal(t, "old_owner_tenant_id", handlerObj.getMetaParentNodeField())
	testifyAssert.Equal(t, "old_follow_tenant", handlerObj.getModeFollowParent())
}

type mockRuleHandler struct {
	returnValue bool
}

func (m *mockRuleHandler) Check(ctx eosc.IEntry) bool {
	return m.returnValue
}

type mockAssertErrorHttpContext struct {
	mockHttpContext
}

func (m *mockAssertErrorHttpContext) Assert(i interface{}) error {
	return strconv.ErrSyntax
}

func TestAccessHierarchy_DoHttpFilterAssertError(t *testing.T) {
	ah := &AccessHierarchy{
		rules: []ruleHandler{&mockRuleHandler{returnValue: false}},
	}
	ctx := &mockAssertErrorHttpContext{}
	chain := &mockChain{}
	err := ah.DoHttpFilter(ctx, chain)
	testifyAssert.Error(t, err)
}

func BenchmarkHandler_Check_Simple(b *testing.B) {
	oldVar := customerVar
	defer func() { customerVar = oldVar }()

	customerVar = &mockCustomerVar{
		data: map[string]map[string]string{
			"app_groups:app1": {
				"group1": "0",
			},
			"group_meta:group1": {
				"mode": "explicit_resources",
			},
			"group_resources:group1": {
				"res1": "0",
			},
		},
	}

	h := &handler{
		a: &mockMetrics{val: "app1"},
		b: &mockMetrics{val: "res1"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h.Check(&mockEntry{})
	}
}

func BenchmarkHandler_Check_10LevelsComplex(b *testing.B) {
	oldVar := customerVar
	defer func() { customerVar = oldVar }()

	mockData := map[string]map[string]string{
		"user_roles:user1": {
			"role_1":        "0",
			"role_1_extra1": "0",
			"role_1_extra2": "0",
			"role_1_extra3": "0",
			"role_1_extra4": "0",
		},
	}

	for i := 1; i <= 9; i++ {
		currRole := "role_" + strconv.Itoa(i)
		nextRole := "role_" + strconv.Itoa(i+1)
		mockData["role_meta:"+currRole] = map[string]string{
			"mode":           "follow",
			"parent_role_id": nextRole,
		}

		for k := 1; k <= 4; k++ {
			extraRole := currRole + "_extra" + strconv.Itoa(k)
			mockData["role_meta:"+extraRole] = map[string]string{
				"mode": "explicit",
			}
			mockData["role_permissions:"+extraRole] = map[string]string{
				"other_permission": "0",
			}
		}
	}

	mockData["role_meta:role_10"] = map[string]string{
		"mode": "explicit",
	}
	mockData["role_permissions:role_10"] = map[string]string{
		"permission1": "0",
	}

	for k := 1; k <= 4; k++ {
		extraRole := "role_10_extra" + strconv.Itoa(k)
		mockData["role_meta:"+extraRole] = map[string]string{
			"mode": "explicit",
		}
	}

	customerVar = &mockCustomerVar{data: mockData}

	ah := &AccessHierarchy{}
	h := ah.newHandler("$ctx_app", "$ctx_res", &HierarchyConfig{
		AppNodesPrefix:      "user_roles:",
		NodeMetaPrefix:      "role_meta:",
		NodeTargetsPrefix:   "role_permissions:",
		MetaModeField:       "mode",
		MetaParentNodeField: "parent_role_id",
		ModeExplicit:        "explicit",
		ModeFollowParent:    "follow",
	})

	handlerObj := h.(*handler)
	handlerObj.a = &mockMetrics{val: "user1"}
	handlerObj.b = &mockMetrics{val: "permission1"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h.Check(&mockEntry{})
	}
}

func BenchmarkHandler_Check_DeepInheritance(b *testing.B) {
	oldVar := customerVar
	defer func() { customerVar = oldVar }()

	customerVar = &mockCustomerVar{
		data: map[string]map[string]string{
			"user_roles:user1": {
				"role1": "0",
			},
			"role_meta:role1": {
				"mode":           "follow",
				"parent_role_id": "role2",
			},
			"role_meta:role2": {
				"mode":           "follow",
				"parent_role_id": "role3",
			},
			"role_meta:role3": {
				"mode": "explicit",
			},
			"role_permissions:role3": {
				"permission1": "0",
			},
		},
	}

	ah := &AccessHierarchy{}
	h := ah.newHandler("$ctx_app", "$ctx_res", &HierarchyConfig{
		AppNodesPrefix:      "user_roles:",
		NodeMetaPrefix:      "role_meta:",
		NodeTargetsPrefix:   "role_permissions:",
		MetaModeField:       "mode",
		MetaParentNodeField: "parent_role_id",
		ModeExplicit:        "explicit",
		ModeFollowParent:    "follow",
	})

	handlerObj := h.(*handler)
	handlerObj.a = &mockMetrics{val: "user1"}
	handlerObj.b = &mockMetrics{val: "permission1"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h.Check(&mockEntry{})
	}
}
