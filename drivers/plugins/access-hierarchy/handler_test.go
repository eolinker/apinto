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

// mockMetrics 实现了 metrics.Metrics 接口，用于在测试中模拟指标解析表达式的返回结果
type mockMetrics struct {
	val string
}

func (m *mockMetrics) Metrics(entry metrics.LabelReader) string {
	return m.val
}

func (m *mockMetrics) Key() string {
	return m.val
}

// mockEntry 模拟实现了 eosc.IEntry 接口，作为测试运行时的参数
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

// mockCustomerVar 模拟实现了 eosc.ICustomerVar 接口，用于保存和查询租户及资源组配置关系
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
		// 场景：AppID 或 ResourceID 提取出的结果为空
		// 预期：无法解析出有效指标，安全拒绝，返回 false
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
		// 场景：App 没有绑定 to 资源组
		// 预期：在 app_groups 映射中无对应记录，返回 false
		mockVar := &mockCustomerVar{data: make(map[string]map[string]string)}
		customerVar = mockVar

		h := &handler{
			a: &mockMetrics{val: "app1"},
			b: &mockMetrics{val: "res1"},
		}
		testifyAssert.False(t, h.Check(&mockEntry{}))
	})

	t.Run("simple explicit resource match pass", func(t *testing.T) {
		// 场景：最基础的显式资源组绑定关系
		// 关系：App1 -> Group1 (显式资源模式) -> 资源1
		// 预期：未过期，直接匹配成功，返回 true
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
		// 场景：应用与资源组的关联关系已过期
		// 关系：App1 -> Group1 (绑定关系中时间戳在当前时间之前)
		// 预期：检测到关系过期，拒绝访问，返回 false
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
		// 场景：资源与资源组的关联关系已过期
		// 关系：Group1 -> 资源1 (绑定关系中时间戳在当前时间之前)
		// 预期：检测到资源授权过期，拒绝访问，返回 false
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
		// 场景：资源组配置为“跟随租户”模式，能够沿着租户继承路径找到所辖资源
		// 关系：App1 -> Group1 (跟随租户 tenant1) -> 租户拥有 Group2 (显式资源模式) -> 资源1
		// 预期：DFS 递归正确匹配该继承关系，返回 true
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
		// 场景：跟随租户模式下，过滤租户下非“显式资源”类型的子组，防止（租户 -> 组 -> 租户）无限递归
		// 关系：App1 -> Group1 (跟随租户 tenant1) -> 租户拥有 Group2 (但 Group2 又是跟随租户模式，非显式资源模式)
		// 预期：跳过该组，不进行进一步递归，校验失败，返回 false
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
		// 场景：极其复杂的跨多级租户、多级被授权资源组（Granted Groups）的深度继承匹配
		// 关系：
		//   1. App1 绑定到 Group1
		//   2. Group1 的元数据表明其跟随租户 tenant_child
		//   3. tenant_child 被授予了（Granted）Group_parent 资源组权限
		//   4. Group_parent 元数据表明其跟随上级租户 tenant_parent
		//   5. tenant_parent 旗下拥有 Group_grandparent 资源组（显式资源模式）
		//   6. Group_grandparent 绑定了目标资源 res1
		// 预期：深层 DFS 递归能完美寻路通过，返回 true
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
		// 场景：防御性测试 - 资源组依赖成环
		// 关系：Group1 -> 租户1 -> Group2 -> 租户2 -> Group1（形成环状依赖）
		// 预期：DFS 递归检测到 Group1 已访问过，安全中断递归，不导致堆栈溢出，最终返回 false
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
		// 场景：防御性测试 - 租户级联依赖成环
		// 关系：Group1 -> 租户1 -> Group2 -> 租户1（形成环状依赖）
		// 预期：DFS 递归检测到 租户1 已访问过，安全中断递归，不导致堆栈溢出，最终返回 false
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
	// 测试 AccessHierarchy 的工作器基本生命周期方法与 Skill 校验
	ah := &AccessHierarchy{}
	testifyAssert.NoError(t, ah.Start())
	testifyAssert.NoError(t, ah.Stop())
	ah.Destroy()

	testifyAssert.True(t, ah.CheckSkill(http_context.FilterSkillName))
	testifyAssert.False(t, ah.CheckSkill("some_other_skill"))
}

func TestAccessHierarchy_AssertAndParse(t *testing.T) {
	// 测试配置格式 of Assert（类型断言转换）和规则解析
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
		ah.newHandler("$ctx_app", "$ctx_res"),
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
	h := ah.newHandler("ctx_app", "ctx_res")
	testifyAssert.NotNil(t, h)
}
