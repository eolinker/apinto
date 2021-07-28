package checker

import (
	"reflect"
	"testing"
)

func TestCreateChecker(t *testing.T) {
	type  checkValue struct {
		v string
		has bool
	}
	type args struct {
		pattern string
		value checkValue
	}
	regexp, _ := newCheckerRegexp("abc")
	regexpG, _ := newCheckerRegexpG("abc")
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "equal",
			args: args{
				pattern: "abc",
				value:   checkValue{
					v:"abc",
					has:true,
				},
			},
			want:   true,
			wantErr: false,
		},
		{
			name: "equal-not",
			args: args{
				pattern: "=abc",
				value:   checkValue{
					v:"abcd",
					has:true,
				},
			},
			want:   false,
			wantErr: false,
		},
		{
			name: "equal2",
			args: args{
				pattern: "=abc",
			},
			want:    newCheckerEqual("abc"),
			wantErr: false,
		}, {
			name:    "all",
			args:    args{
				pattern: "*",
			},
			want:    newCheckerAll(),
			wantErr: false,
		}, {
			name: "exist",
			args: args{
				pattern: "**",
			},
			want:    newCheckerExist(),
			wantErr: false,
		}, {
			name: "not exist",
			args: args{
				pattern: "!",
			},
			want:    newCheckerNotExits(),
			wantErr: false,
		},{
			name: "none",
			args: args{
				pattern: "$",
			},
			want:    newCheckerNone(),
			wantErr: false,
		}, {
			name:    "not equal",
			args:    args{
				pattern: "!=abc",
			},
			want:    newCheckerNotEqual("abc"),
			wantErr: false,
		},
		{
			name:    "prefix",
			args:    args{
				pattern: "^=/abc",
			},
			want:    newCheckerPrefix("/abc"),
			wantErr: false,
		},{
			name:    "regex",
			args:    args{
				pattern: "~=/abc/",
			},
			want:   regexp ,
			wantErr: false,
		},
		{
			name:    "regex global",
			args:    args{
				pattern: "~* =/abc/",
			},
			want:   regexpG ,
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
			got:=checker.Check(tt.args.value.v,tt.args.value.has)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() got = %v, want %v", got, tt.want)
			}
		})
	}
}