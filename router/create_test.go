package router

import (
	"github.com/eolinker/goku-eosc/router/checker"
	"strings"
	"testing"
)

type testSource map[string]string

func (t testSource) Get(cmd string) (string, bool) {
	v,has:=t[cmd]
	return v,has
}

var (
	testSourcesList = []testSource{
		{
			"location":"/abc",
			"header:a":"10",
		},
		{
			"location":"/abc",
			"header:a":"1",
		},
	}
)


type TestHelper struct {
	index map[string]int
}

func NewTestHelper(cmds []string) *TestHelper {
	index:=make(map[string]int)
	for i,cmd:=range cmds{
		index[cmd] = i
	}
	return &TestHelper{index:index}
}

func (t *TestHelper) Less(i, j string) bool {
	cmdI,keyI := t.cmdType(i)
	cmdJ,keyJ:=t.cmdType(j)
	if cmdI != cmdJ{
		 ii,hasI:=t.index[cmdI]
		 jj,hasJ:=t.index[cmdJ]
		 if !hasI && !hasJ{
		 	 return cmdI<cmdJ
		 }
		 if !hasJ{
		 	return true
		 }
		 if !hasI{
		 	return  false
		 }
		 return ii<jj
	}
	return keyI<keyJ
}
func (t *TestHelper)cmdType(cmd string)(string ,string) {
	i:=strings.Index(cmd,":")
	if i < 0{
		return cmd,""
	}
	if i == 0{
		return strings.ToLower(cmd[1:]),""
	}

	return strings.ToLower(cmd[:i]),strings.ToLower(cmd[i+1:])

}

type TestRule struct {
	paths []string
	target string
}

func (tr *TestRule) toRule()Rule  {
	path:=make([]RulePath,0,len(tr.paths))
	for _,p:=range tr.paths{
		i:=strings.Index(p," ")
		if i<0{
			continue
		}
		c,e:=checker.Parse(p[i:])
		if e!=nil{
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
func TestParseRouter(t *testing.T) {


	helper := NewTestHelper([]string{"location","header","query"})
	tests := []struct {
		name    string
		args    []*TestRule
		want    []string
		wantErr bool
	}{
		{
			name: "demo",
			args: []*TestRule{
				{
					paths:  []string{"location = /abc","header:a = 10"},
					target: "demo",
				},
				{
					paths:  []string{"location = /abc","header:a != 10"},
					target: "demo2",
				},
			},
			want:    []string{"demo","demo2"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules :=make([]Rule,0,len(tt.args))
			for _,r:=range tt.args{
				rules = append(rules, r.toRule())
			}
			r, err := ParseRouter(rules, helper)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRouter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for i,s:=range testSourcesList{
				target,_:=r.Router(s)
				if tt.want[i] != target{
					t.Errorf("router(sources[%d]) got = %v, want %s",i, target, tt.want[i])
				}else {
					t.Logf("router(sources[%d]) got = \"%v\", ok",i, target)
				}

			}
		})
	}
}