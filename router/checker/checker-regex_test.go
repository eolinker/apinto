/*
 * Copyright (c) 2021. Lorem ipsum dolor sit amet, consectetur adipiscing elit.
 * Morbi non lorem porttitor neque feugiat blandit. Ut vitae ipsum eget quam lacinia accumsan.
 * Etiam sed turpis ac ipsum condimentum fringilla. Maecenas magna.
 * Proin dapibus sapien vel ante. Aliquam erat volutpat. Pellentesque sagittis ligula eget metus.
 * Vestibulum commodo. Ut rhoncus gravida arcu.
 */

package checker

import (
	"testing"
)

func Test_checkerRegexp_Check(t1 *testing.T) {

	type args struct {
		v   string
		has bool
	}
	tests := []struct {
		name   string
		pattern string
		args   args
		want   bool
	}{
		{
			name:    "size",
			pattern: "[a-z]{1,10}",
			args: args{
				v:   "a",
				has: true,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t,err := newCheckerRegexpG(tt.pattern)
			if err!= nil{
				t1.Errorf("parse checker Regexp () error:%v, not want error",err)
				return
			}
			if got := t.Check(tt.args.v, tt.args.has); got != tt.want {
				t1.Errorf("Check() = %v, want %v", got, tt.want)
			}
		})
	}
}