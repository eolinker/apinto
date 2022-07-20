package router

import (
	"strings"
	"testing"

	http_service "github.com/eolinker/eosc/context/http-context"
)

type testSource map[string]string

func (t testSource) Get(cmd string) (string, bool) {
	v, has := t[cmd]
	return v, has
}

type TestHelper struct {
	index map[string]int
}

func NewTestHelper(cmds []string) *TestHelper {
	index := make(map[string]int)
	for i, cmd := range cmds {
		index[cmd] = i
	}
	return &TestHelper{index: index}
}

func (t *TestHelper) Less(i, j string) bool {
	cmdI, keyI := t.cmdType(i)
	cmdJ, keyJ := t.cmdType(j)
	if cmdI != cmdJ {
		ii, hasI := t.index[cmdI]
		jj, hasJ := t.index[cmdJ]
		if !hasI && !hasJ {
			return cmdI < cmdJ
		}
		if !hasJ {
			return true
		}
		if !hasI {
			return false
		}
		return ii < jj
	}
	return keyI < keyJ
}
func (t *TestHelper) cmdType(cmd string) (string, string) {
	i := strings.Index(cmd, ":")
	if i < 0 {
		return cmd, ""
	}
	if i == 0 {
		return strings.ToLower(cmd[1:]), ""
	}

	return strings.ToLower(cmd[:i]), strings.ToLower(cmd[i+1:])

}

type TestRule struct {
	paths  []string
	target string
}

func (tr *TestRule) toRule() Rule {
	path := make([]RulePath, 0, len(tr.paths))
	for _, p := range tr.paths {
		i := strings.Index(p, " ")
		if i < 0 {
			continue
		}
		c, e := http_service.Parse(p[i:])
		if e != nil {
			continue
		}
		path = append(path, RulePath{
			CMD:     p[:i],
			Checker: c,
		})
	}
	return Rule{
		Path:   path,
		Target: tr.target,
	}
}

var tests = []struct {
	name     string
	args     []*TestRule
	want     []string
	testCase []testSource
	wantErr  bool
}{
	{
		name: "检测互斥",
		testCase: []testSource{
			{
				"location": "/abc",
				"header:a": "10",
			},
			{
				"location": "/abc",
				"header:a": "1",
			},
		},
		args: []*TestRule{
			{
				paths:  []string{"location = /abc", "header:a = 10"},
				target: "demo",
			},
			{
				paths:  []string{"location = /abc", "header:a != 10"},
				target: "demo2",
			},
		},
		want:    []string{"demo", "demo2"},
		wantErr: false,
	},
	{
		name: "路由树有多余节点2",
		testCase: []testSource{
			{
				"location":    "/abc",
				"header:a":    "1",
				"query:name":  "liu",
				"query:phone": "13312313412",
			},
			{
				"location":    "/abc",
				"header:a":    "1",
				"query:phone": "133123134",
			},
			{
				"location":    "/abc",
				"header:a":    "1",
				"query:phone": "133123",
			},
			{
				"location": "/abc",
				"header:a": "1",
			},
		},
		args: []*TestRule{
			{
				paths:  []string{"location = /abc", "header:a != 10", "query:name = liu", "query:phone = 13312313412"},
				target: "demo",
			},
			{
				paths:  []string{"location = /abc", "header:a != 10", "query:phone ^= 133123"},
				target: "demo2",
			},
			{
				paths:  []string{"location = /abc", "header:a != 10", "query:phone = 133123"},
				target: "demo3",
			},
			{
				paths:  []string{"location = /abc", "header:a != 10"},
				target: "demo4",
			},
		},
		want:    []string{"demo", "demo2", "demo3", "demo4"},
		wantErr: false,
	},
	{
		name: "测试前缀匹配",
		testCase: []testSource{
			{
				"location": "/abc/adw",
				"header:a": "1",
			},
			{
				"location": "/abc",
				"header:a": "1",
			},
			{
				"location": "/abcdasdwq",
				"header:a": "1",
			},
		},
		args: []*TestRule{
			{
				paths:  []string{"location ^= /abc/", "header:a != 10"},
				target: "demo",
			},
			{
				paths:  []string{"location = /abc", "header:a != 10"},
				target: "demo2",
			},
			{
				paths:  []string{"location ^= /abc", "header:a != 10"},
				target: "demo3",
			},
		},
		want:    []string{"demo", "demo2", "demo3"},
		wantErr: false,
	},
	{
		name: "测试匹配类型优先级1",
		testCase: []testSource{
			{
				"location": "/abc",
				"header:a": "a",
			},
			{
				"location": "/abc",
				"header:a": "A",
			},
			{
				"location": "/abc",
				"header:a": "aa",
			},
			{
				"location": "/abc",
				"header:a": "aA",
			},
			{
				"location": "/abc",
				"header:a": "Aa",
			},
		},
		args: []*TestRule{
			{
				paths:  []string{"location = /abc", "header:a = a"},
				target: "demo",
			},
			{
				paths:  []string{"location = /abc", "header:a != a"},
				target: "demo2",
			},
			{
				paths:  []string{"location = /abc", "header:a ~= [a-z]{1,10}"},
				target: "demo3",
			},
			{
				paths:  []string{"location = /abc", "header:a ~*= [a-z]{1,10}"},
				target: "demo4",
			},
		},
		want:    []string{"demo", "demo2", "demo2", "demo2", "demo2"},
		wantErr: false,
	},
	{
		name: "测试匹配类型优先级2",
		testCase: []testSource{
			{
				"location": "/abc",
				"header:a": "a",
			},
			{
				"location": "/abc",
				"header:a": "A",
			},
			{
				"location": "/abc",
				"header:b": "a",
			},
			{
				"location": "/abc",
				"header:b": "aA",
			},
			{
				"location": "/abc",
				"header:b": "Aa",
			},
		},
		args: []*TestRule{
			{
				paths:  []string{"location = /abc", "header:a = a"},
				target: "demo",
			},
			{
				paths:  []string{"location = /abc", "header:a != a"},
				target: "demo2",
			},
			{
				paths:  []string{"location = /abc", "header:b ~= ^[a-z]{1,10}$"},
				target: "demo3",
			},
			{
				paths:  []string{"location = /abc", "header:b ~*= ^[a-z]{1,10}$"},
				target: "demo4",
			},
		},
		want:    []string{"demo", "demo2", "demo3", "demo4", "demo4"},
		wantErr: false,
	},
	{
		name: "存在和不存在互斥用例",
		testCase: []testSource{
			{
				"location": "/abc",
				"header:a": "a",
			},
			{
				"location": "/abc",
				"header:a": "A",
			},
			{
				"location": "/abc",
				"header:b": "a",
			},
		},
		args: []*TestRule{
			{
				paths:  []string{"location = /abc", "header:a **"},
				target: "demo",
			},
			{
				paths:  []string{"location = /abc", "header:a = A"},
				target: "demo2",
			},
			{
				paths:  []string{"location = /abc", "header:a !"},
				target: "demo3",
			},
		},
		want:    []string{"demo", "demo2", "demo3"},
		wantErr: false,
	},
	{
		name: "检测前缀",
		testCase: []testSource{
			{
				"location":    "/abc",
				"header:a":    "1",
				"query:name":  "liu",
				"query:phone": "13312313412",
			},
			{
				"location":    "/abc",
				"header:a":    "1",
				"query:name":  "liu",
				"query:phone": "13312313412",
				"query:mail":  "demo@eolinker.com",
			},
			{
				"location":    "/abc",
				"header:a":    "1",
				"query:name":  "liu",
				"query:phone": "13312313412",
				"query:mail":  "demoabc",
			},
		},
		args: []*TestRule{
			{
				paths:  []string{"location = /abc", "header:a != 10", "query:name = liu", "query:phone = 13312313412"},
				target: "demo",
			},
			{
				paths:  []string{"location = /abc", "header:a != 10", "query:name = liu", "query:mail = demo@eolinker.com"},
				target: "demo2",
			},
			{
				paths:  []string{"location = /abc", "header:a != 10", "query:name = liu", "query:mail ^= demo"},
				target: "demo3",
			},
		},
		want:    []string{"demo", "demo2", "demo3"},
		wantErr: false,
	},
	{
		name: "检测重复路由规则覆盖",
		testCase: []testSource{
			{
				"location":    "/abc",
				"header:a":    "1",
				"query:name":  "liu",
				"query:phone": "13312313412",
				"query:mail":  "demo@eolinker.com",
			},
			{
				"location":    "/abc",
				"header:a":    "1",
				"query:name":  "liu",
				"query:phone": "13312313412",
				"query:mail":  "demoabc",
			},
		},
		args: []*TestRule{
			{
				paths:  []string{"query:name = liu", "query:phone = 13312313412", "query:mail = demo@eolinker.com"},
				target: "demo",
			},
			{
				paths:  []string{"query:name = liu", "query:mail = demo@eolinker.com", "query:phone = 13312313412"},
				target: "demo2",
			},
			{
				paths:  []string{"query:name = liu", "query:mail ^= demo"},
				target: "demo3",
			},
		},
		want:    []string{"demo2", "demo3"},
		wantErr: true,
	},
	{
		name: "复杂检测1",
		testCase: []testSource{
			{
				"host":       "a.abc.com",
				"location":   "",
				"header:a":   "1",
				"query:name": "liu",
			},
			{
				"host":       "a.abc.com",
				"location":   "",
				"header:a":   "10",
				"query:name": "liu",
			},
		},
		args: []*TestRule{
			{
				paths:  []string{"host = a.abc.com ", "location $", "header:a = 10", "query:name = chen"},
				target: "demo1",
			},
			{
				paths:  []string{"host = a.abc.com ", "location $", "header:a != 10", "query:name = liu"},
				target: "demo2",
			},
			{
				paths:  []string{"host ^= a.abc", "location $", "header:a != 10", "query:name = chen"},
				target: "demo3",
			},
			{
				paths:  []string{"host ^= a.abc", "location $", "header:a = 10", "query:name = liu"},
				target: "demo4",
			},
		},
		want:    []string{"demo2", "demo4"},
		wantErr: false,
	},
	{
		name: "检测匹配路径不存在的情况",
		testCase: []testSource{
			{
				"host":       "a.abc.com",
				"header:a":   "1",
				"query:name": "wu",
			},
			{
				"host":       "a.abc.com",
				"header:a":   "10",
				"query:name": "wu",
			},
		},
		args: []*TestRule{
			{
				paths:  []string{"host = a.abc.com ", "header:a = 10", "query:name = chen"},
				target: "demo1",
			},
			{
				paths:  []string{"host = a.abc.com ", "header:a != 10", "query:name = liu"},
				target: "demo2",
			},
			{
				paths:  []string{"host ^= a.abc", "header:a != 10", "query:name = chen"},
				target: "demo3",
			},
			{
				paths:  []string{"host ^= a.abc", "header:a = 10", "query:name = liu"},
				target: "demo4",
			},
		},
		want:    []string{"", ""},
		wantErr: false,
	},
	{
		name: "测试任意",
		testCase: []testSource{
			{
				"host":       "a.abc.com",
				"header:a":   "1",
				"header:b":   "bbb",
				"query:name": "chen",
			},
			{
				"host":       "a.abc.com",
				"header:a":   "10",
				"header:b":   "ccc",
				"query:name": "liu",
			},
		},
		args: []*TestRule{
			{
				paths:  []string{"host = a.abc.com ", "header:b = ", "header:a = 10", "query:name = chen"},
				target: "demo1",
			},
			{
				paths:  []string{"host = a.abc.com ", "header:b =* ", "header:a != 10", "query:name = liu"},
				target: "demo2",
			},
			{
				paths:  []string{"host ^= a.abc", "header:b =", "header:a != 10", "query:name = chen"},
				target: "demo3",
			},
			{
				paths:  []string{"host ^= a.abc", "header:b * ", "header:a = 10", "query:name = liu"},
				target: "demo4",
			},
		},
		want:    []string{"demo3", "demo4"},
		wantErr: false,
	},
	{
		name: "检测同时匹配多个优先",
		testCase: []testSource{
			{
				"host":       "a.abc.com",
				"header:a":   "aaa",
				"header:b":   "bbb",
				"header:c":   "ccc",
				"query:name": "chen",
			},
			{
				"host":       "a.abc.com",
				"header:a":   "aaa",
				"header:b":   "bbb",
				"query:name": "chen",
			},
			{
				"host":       "a.abc.com",
				"header:a":   "aaa",
				"query:name": "chen",
			},
		},
		args: []*TestRule{
			{
				paths:  []string{"host = a.abc.com", "header:a = aaa", "query:name = chen"},
				target: "demo1",
			},
			{
				paths:  []string{"host = a.abc.com ", "query:name = chen", "header:b = bbb ", "header:a = aaa"},
				target: "demo2",
			},
			{
				paths:  []string{"host = a.abc.com ", "header:c = ccc ", "query:name = chen", "header:b = bbb", "header:a = aaa"},
				target: "demo3",
			},
			{
				paths:  []string{"host = a.abc.com", "header:a = aaa", "query:name = chen", "query:xxx = chen"},
				target: "demo4",
			},
		},
		want:    []string{"demo3", "demo2", "demo1"},
		wantErr: false,
	},
	{
		name: "检测路径上有重复指标",
		testCase: []testSource{
			{
				"host":       "a.abc.com",
				"query:name": "chen",
			},
		},
		args: []*TestRule{
			{
				paths:  []string{"host = a.abc.com", "host = a.abc.com", "query:name = chen"},
				target: "demo1",
			},
		},
		want:    []string{"demo1"},
		wantErr: false,
	},
	{
		name: "检测匹配路由时，匹配规则的优先级优于满足多个条件的优先级",
		testCase: []testSource{
			{
				"host":       "a.abc.com",
				"location":   "/abc",
				"query:name": "chen",
			},
		},
		args: []*TestRule{
			{
				paths:  []string{"host = a.abc.com", "location = /abc"},
				target: "demo1",
			},
			{
				paths:  []string{"host ^= a.abc", "location = /abc", "query:name = chen"},
				target: "demo2",
			},
		},
		want:    []string{"demo1"},
		wantErr: false,
	},
	{
		name: "测试method",
		testCase: []testSource{
			{
				"host":     "a.abc.com",
				"location": "/abc",
				"method":   "GET",
			},
			{
				"host":     "a.abc.com",
				"location": "/abc",
				"method":   "POST",
			},
		},
		args: []*TestRule{
			{
				paths:  []string{"host = a.abc.com", "location = /abc", "method = GET"},
				target: "demo1",
			},
			{
				paths:  []string{"host = a.abc.com", "location = /abc", "method = POST"},
				target: "demo2",
			},
			{
				paths:  []string{"host = a.abc.com.cn", "location = /abc", "method = GET"},
				target: "demo3",
			},
			{
				paths:  []string{"host = a.abc.com.cn", "location = /abc", "method = POST"},
				target: "demo4",
			},
		},

		want:    []string{"demo1", "demo2"},
		wantErr: false,
	},
	{
		name: "测试匹配优先级规则排序",
		testCase: []testSource{
			{
				"host":     "a.abc.com",
				"location": "/abc",
				"method":   "GET",
			},
		},
		args: []*TestRule{
			{
				paths:  []string{"host ^=*a.abc.", "location = /abc", "method = GET"},
				target: "demo1",
			},
			{
				paths:  []string{"host != a.abc.", "location = /abc", "method = GET"},
				target: "demo2",
			},
			{
				paths:  []string{"host ^= a.abc.com", "location = /abc", "method = GET"},
				target: "demo3",
			},
			{
				paths:  []string{"host ^= a.abc.com.cn", "location = /abc", "method = GET"},
				target: "demo4",
			},
		},

		want:    []string{"demo3"},
		wantErr: false,
	},
}

func TestParseRouter(t *testing.T) {

	helper := NewTestHelper([]string{"host", "method", "location", "header", "query"})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := make([]Rule, 0, len(tt.args))
			for _, r := range tt.args {
				rules = append(rules, r.toRule())
			}
			r, err := ParseRouter(rules, helper)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRouter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				t.Logf("ParseRouter() error = %v  ok", err)
				return
			}

			for i, s := range tt.testCase {
				endpoint, h := r.Router(s)
				target := ""
				if h {
					target = endpoint.Target()
				}
				if tt.want[i] != target {
					t.Errorf("router(sources[%d]) got = %v, want %s", i, target, tt.want[i])
				} else {
					t.Logf("router(sources[%d]) got = \"%v\", ok", i, target)

				}

			}
		})
	}
}
