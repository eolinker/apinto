package params_check_v2

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eolinker/apinto/checker"
	"github.com/ohler55/ojg/oj"
)

// MockHeaderReader 模拟 header 读取器
type MockHeaderReader struct {
	headers map[string]string
}

func (m *MockHeaderReader) RawHeader() string {
	//TODO implement me
	panic("implement me")
}

func (m *MockHeaderReader) Headers() http.Header {
	//TODO implement me
	panic("implement me")
}

func (m *MockHeaderReader) Host() string {
	//TODO implement me
	panic("implement me")
}

func (m *MockHeaderReader) GetCookie(key string) string {
	//TODO implement me
	panic("implement me")
}

func (m *MockHeaderReader) GetHeader(name string) string {
	return m.headers[name]
}

// MockQueryReader 模拟 query 读取器
type MockQueryReader struct {
	queries map[string]string
}

func (m *MockQueryReader) RawQuery() string {
	//TODO implement me
	panic("implement me")
}

func (m *MockQueryReader) GetQuery(name string) string {
	return m.queries[name]
}

func TestParamCheckLogic(t *testing.T) {
	tests := []testStruct{
		{
			name: "And logic",
			param: &Param{
				Logic: logicAnd,
				Params: []*SubParam{
					{
						Position:  positionHeader,
						Name:      "X-Test-Header",
						MatchText: "test-value",
					},
					{
						Position:  positionQuery,
						Name:      "query-param",
						MatchText: "query-value",
					},
				},
			},
			header: map[string]string{
				"X-Test-Header": "test-value",
			},
			query: map[string]string{
				"query-param": "query-value",
			},
			expected:    true,
			expectError: false,
		},
		{
			name: "Or logic",
			param: &Param{
				Logic: logicOr,
				Params: []*SubParam{
					{
						Position:  positionHeader,
						Name:      "X-Test-Header",
						MatchText: "test-value",
					},
					{
						Position:  positionQuery,
						Name:      "query-param",
						MatchText: "query-value",
					},
				},
			},
			header: map[string]string{
				"X-Test-Header": "test-value",
			},
			query: map[string]string{
				"query-param": "query-value",
			},
			expected:    true,
			expectError: false,
		},
		{
			name: "And logic (fail case)",
			param: &Param{
				Logic: logicAnd,
				Params: []*SubParam{
					{
						Position:  positionHeader,
						Name:      "X-Test-Header",
						MatchText: "test-value",
					},
					{
						Position:  positionQuery,
						Name:      "query-param",
						MatchText: "query-value",
					},
				},
			},
			header: map[string]string{
				"X-Test-Header": "test-value",
			},
			query: map[string]string{
				"query-param": "query-value-2",
			},
			expected:    false,
			expectError: false,
		},
		{
			name: "Or logic (fail case)",
			param: &Param{
				Logic: logicOr,
				Params: []*SubParam{
					{
						Position:  positionHeader,
						Name:      "X-Test-Header",
						MatchText: "test-value",
					},
					{
						Position:  positionQuery,
						Name:      "query-param",
						MatchText: "query-value",
					},
				},
			},
			header: map[string]string{
				"X-Test-Header": "test-value-2",
			},
			query: map[string]string{
				"query-param": "query-value-2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ck, err := newParamChecker(tt.param)
			assert.NoError(t, err)
			assert.NotNil(t, ck)
			assert.Equal(t, tt.expected, ck.Check(&MockHeaderReader{headers: tt.header}, &MockQueryReader{queries: tt.query}, nil, nil))
		})
	}
}

type testStruct struct {
	name        string
	param       *Param
	header      map[string]string
	query       map[string]string
	body        string
	expected    bool
	expectError bool
}

func TestParamCheck(t *testing.T) {
	data := `{
		"code":"HWJXH_DEPTCOMPLEX_CSGLJ_HWJXHXT_XLJXHZYL_1.0",
		"isPage":true,
		"index":1,
		"size":10,
		"apiType":"deptCOMPLEX",
		"userName":"APIFXJCXT",
		"psd":"5190649892064f8c9bf387c3d30ce021",
		"apiId":"d1d3f080f18448fcb5354f67b93e54e9",
		"search":[
			{"param":"F_SECTIONNAME","type":"String","val":"value1"},
			{"param":"F_SECTIONNAME","type":"String","val":""}
		]
	}`

	tests := []struct {
		name        string
		param       *Param
		header      map[string]string
		query       map[string]string
		expected    bool
		expectError bool
	}{
		{
			name: "Body array match any",
			param: &Param{
				Params: []*SubParam{
					{
						Position:  positionBody,
						Name:      "search[*].val",
						MatchText: "**",
						MatchMode: checker.JsonArrayMatchAny,
					},
					{
						Position:  positionBody,
						Name:      "search[1].val",
						MatchText: "**",
						MatchMode: checker.JsonArrayMatchAny,
					},
				},
				Logic: logicAnd,
			},
			expected:    true,
			expectError: false,
		},
		{
			name: "Header match",
			param: &Param{
				Params: []*SubParam{
					{
						Position:  positionHeader,
						Name:      "X-Test-Header",
						MatchText: "test-value",
					},
				},
				Logic: logicAnd,
			},
			header: map[string]string{
				"X-Test-Header": "test-value",
			},
			expected:    true,
			expectError: false,
		},
		{
			name: "Query match",
			param: &Param{
				Params: []*SubParam{
					{
						Position:  positionQuery,
						Name:      "query-param",
						MatchText: "query-value",
					},
				},
				Logic: logicAnd,
			},
			query: map[string]string{
				"query-param": "query-value",
			},
			expected:    true,
			expectError: false,
		},
		{
			name: "Body array match all (fail case)",
			param: &Param{
				Params: []*SubParam{
					{
						Position:  positionBody,
						Name:      "search[*].val",
						MatchText: "**",
						MatchMode: checker.JsonArrayMatchAll,
					},
				},
				Logic: logicAnd,
			},
			expected:    false,
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// 解析 body 数据
			body, err := oj.ParseString(data)
			if err != nil {
				t.Fatalf("failed to parse body: %v", err)
			}

			// 创建 paramChecker
			ck, err := newParamChecker(test.param)
			if err != nil {
				if !test.expectError {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			// 创建 header 和 query 读取器
			headerReader := &MockHeaderReader{headers: test.header}
			queryReader := &MockQueryReader{queries: test.query}

			// 执行校验
			ok := ck.Check(headerReader, queryReader, body, jsonChecker)

			// 检查结果
			if ok != test.expected {
				t.Fatalf("expected %v, got %v", test.expected, ok)
			}
		})
	}
}
