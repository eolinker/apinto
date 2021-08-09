package checker

import (
	"reflect"
	"testing"
)

func TestCreateChecker(t *testing.T) {
	type args struct {
		pattern string
	}
	type valueSuccess struct {
		v          string
		has        bool
		wantResult bool
	}
	type valueFail struct {
		v          string
		has        bool
		wantResult bool
	}
	regexp, _ := newCheckerRegexp("^[a-z]{1,10}$")
	regexpG, _ := newCheckerRegexpG("^[a-z]{1,10}$")
	tests := []struct {
		name    string
		args    args
		vs      valueSuccess
		vf      valueFail
		want    Checker
		wantErr bool
	}{
		{
			name: "全等匹配=str",
			args: args{
				pattern: "=abc",
			},
			vs: valueSuccess{
				v:          "abc",
				has:        true,
				wantResult: true,
			},
			vf: valueFail{
				v:          "ab",
				has:        true,
				wantResult: false,
			},
			want:    newCheckerEqual("abc"),
			wantErr: false,
		}, {
			name: "全等匹配=str(=省略",
			args: args{
				pattern: "abc",
			},
			vs: valueSuccess{
				v:          "abc",
				has:        true,
				wantResult: true,
			},
			vf: valueFail{
				v:          "ab",
				has:        true,
				wantResult: false,
			},
			want:    newCheckerEqual("abc"),
			wantErr: false,
		}, {
			name: "任意匹配=",
			args: args{
				pattern: "=",
			},
			vs: valueSuccess{
				v:          "abc",
				has:        true,
				wantResult: true,
			},
			vf: valueFail{
				v:          "abc",
				has:        true,
				wantResult: true,
			},
			want:    newCheckerAll(),
			wantErr: false,
		}, {
			name: "任意匹配=(=省略",
			args: args{
				pattern: "",
			},
			vs: valueSuccess{
				v:          "abc",
				has:        true,
				wantResult: true,
			},
			vf: valueFail{
				v:          "abc",
				has:        true,
				wantResult: true,
			},
			want:    newCheckerAll(),
			wantErr: false,
		}, {
			name: "任意匹配=*",
			args: args{
				pattern: "=*",
			},
			vs: valueSuccess{
				v:          "abc",
				has:        true,
				wantResult: true,
			},
			vf: valueFail{
				v:          "abc",
				has:        true,
				wantResult: true,
			},
			want:    newCheckerAll(),
			wantErr: false,
		}, {
			name: "任意匹配=*(=省略",
			args: args{
				pattern: "*",
			},
			vs: valueSuccess{
				v:          "abc",
				has:        true,
				wantResult: true,
			},
			vf: valueFail{
				v:          "abc",
				has:        true,
				wantResult: true,
			},
			want:    newCheckerAll(),
			wantErr: false,
		}, {
			name: "存在匹配=**",
			args: args{
				pattern: "=**",
			},
			vs: valueSuccess{
				v:          "abc",
				has:        true,
				wantResult: true,
			},
			vf: valueFail{
				v:          "",
				has:        true,
				wantResult: false,
			},
			want:    newCheckerExist(),
			wantErr: false,
		}, {
			name: "存在匹配=**(=省略",
			args: args{
				pattern: "**",
			},
			vs: valueSuccess{
				v:          "abc",
				has:        true,
				wantResult: true,
			},
			vf: valueFail{
				v:          "",
				has:        true,
				wantResult: false,
			},
			want:    newCheckerExist(),
			wantErr: false,
		}, {
			name: "不存在匹配=!",
			args: args{
				pattern: "=!",
			},
			vs: valueSuccess{
				v:          "abc",
				has:        false,
				wantResult: true,
			},
			vf: valueFail{
				v:          "abc",
				has:        true,
				wantResult: false,
			},
			want:    newCheckerNotExits(),
			wantErr: false,
		}, {
			name: "不存在匹配=!(=省略",
			args: args{
				pattern: "!",
			},
			vs: valueSuccess{
				v:          "abc",
				has:        false,
				wantResult: true,
			},
			vf: valueFail{
				v:          "abc",
				has:        true,
				wantResult: false,
			},
			want:    newCheckerNotExits(),
			wantErr: false,
		}, {
			name: "空值匹配=$",
			args: args{
				pattern: "=$",
			},
			vs: valueSuccess{
				v:          "",
				has:        true,
				wantResult: true,
			},
			vf: valueFail{
				v:          "abc",
				has:        true,
				wantResult: false,
			},
			want:    newCheckerNone(),
			wantErr: false,
		}, {
			name: "空值匹配=$(=省略",
			args: args{
				pattern: "$",
			},
			vs: valueSuccess{
				v:          "",
				has:        true,
				wantResult: true,
			},
			vf: valueFail{
				v:          "abc",
				has:        true,
				wantResult: false,
			},
			want:    newCheckerNone(),
			wantErr: false,
		}, {
			name: "不等于匹配!=",
			args: args{
				pattern: "!=abc",
			},
			vs: valueSuccess{
				v:          "ab",
				has:        true,
				wantResult: true,
			},
			vf: valueFail{
				v:          "abc",
				has:        true,
				wantResult: false,
			},
			want:    newCheckerNotEqual("abc"),
			wantErr: false,
		}, {
			name: "前缀匹配^=str",
			args: args{
				pattern: "^=/abc",
			},
			vs: valueSuccess{
				v:          "/abcd",
				has:        true,
				wantResult: true,
			},
			vf: valueFail{
				v:          "abc",
				has:        true,
				wantResult: false,
			},
			want:    newCheckerPrefix("/abc"),
			wantErr: false,
		}, {
			name: "前缀匹配=str*",
			args: args{
				pattern: "=/abc*",
			},
			vs: valueSuccess{
				v:          "/abcd",
				has:        true,
				wantResult: true,
			},
			vf: valueFail{
				v:          "abc",
				has:        true,
				wantResult: false,
			},
			want:    newCheckerPrefix("/abc"),
			wantErr: false,
		}, {
			name: "前缀匹配=str*(=省略",
			args: args{
				pattern: "/abc*",
			},
			vs: valueSuccess{
				v:          "/abcd",
				has:        true,
				wantResult: true,
			},
			vf: valueFail{
				v:          "abc",
				has:        true,
				wantResult: false,
			},
			want:    newCheckerPrefix("/abc"),
			wantErr: false,
		}, {
			name: "后缀匹配^=*str",
			args: args{
				pattern: "^=*abc/",
			},
			vs: valueSuccess{
				v:          "dabc/",
				has:        true,
				wantResult: true,
			},
			vf: valueFail{
				v:          "abc",
				has:        true,
				wantResult: false,
			},
			want:    newCheckerSuffix("abc/"),
			wantErr: false,
		}, {
			name: "后缀匹配=*str",
			args: args{
				pattern: "=*abc/",
			},
			vs: valueSuccess{
				v:          "dabc/",
				has:        true,
				wantResult: true,
			},
			vf: valueFail{
				v:          "abc",
				has:        true,
				wantResult: false,
			},
			want:    newCheckerSuffix("abc/"),
			wantErr: false,
		}, {
			name: "后缀匹配=*str(=省略",
			args: args{
				pattern: "*abc/",
			},
			vs: valueSuccess{
				v:          "dabc/",
				has:        true,
				wantResult: true,
			},
			vf: valueFail{
				v:          "abc",
				has:        true,
				wantResult: false,
			},
			want:    newCheckerSuffix("abc/"),
			wantErr: false,
		}, {
			name: "子串匹配=*str*",
			args: args{
				pattern: "=*abc*",
			},
			vs: valueSuccess{
				v:          "abc",
				has:        true,
				wantResult: true,
			},
			vf: valueFail{
				v:          "adc",
				has:        true,
				wantResult: false,
			},
			want:    newCheckerSub("abc"),
			wantErr: false,
		}, {
			name: "子串匹配=*str*(=省略",
			args: args{
				pattern: "*abc*",
			},
			vs: valueSuccess{
				v:          "abc",
				has:        true,
				wantResult: true,
			},
			vf: valueFail{
				v:          "adc",
				has:        true,
				wantResult: false,
			},
			want:    newCheckerSub("abc"),
			wantErr: false,
		}, {
			name: "正则匹配（区分大小写）",
			args: args{
				pattern: "~=^[a-z]{1,10}$",
			},
			vs: valueSuccess{
				v:          "abc",
				has:        true,
				wantResult: true,
			},
			vf: valueFail{
				v:          "ABc",
				has:        true,
				wantResult: false,
			},
			want:    regexp,
			wantErr: false,
		}, {
			name: "正则匹配（不区分大小写）",
			args: args{
				pattern: "~*=^[a-z]{1,10}$",
			},
			vs: valueSuccess{
				v:          "ABC",
				has:        true,
				wantResult: true,
			},
			vf: valueFail{
				v:          "123",
				has:        true,
				wantResult: false,
			},
			want:    regexpG,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker, err := Parse(tt.args.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(checker, tt.want) {
				t.Errorf("Parse() got = %v, want %v", checker, tt.want)
			}
			//验证check
			if checker != nil {
				//测试成功情况
				checkRes := checker.Check(tt.vs.v, tt.vs.has)
				if checkRes != tt.vs.wantResult {
					t.Errorf("Check() got = %v, want %v", checkRes, tt.vs.wantResult)
				}
				//测试失败情况
				checkRes = checker.Check(tt.vf.v, tt.vf.has)
				if checkRes != tt.vf.wantResult {
					t.Errorf("Check() got = %v, want %v", checkRes, tt.vf.wantResult)
				}
			}
		})
	}
}
