package http_router

import (
	"strings"
	"testing"

	"github.com/eolinker/apinto/checker"

	"github.com/ohler55/ojg/jp"
)

func TestJsonChecker(t *testing.T) {
	jsonData := `{
		"name": "John",
		"friends": [
			{"name": "Alice", "age": 25},
			{"name": "Bob", "age": 28}
		],
		"address": {
			"city": "New York",
			"zipcode": "10001"
		}
	}`
	type args struct {
		name string
		want string
	}
	tests := []args{
		{
			name: "name",
			want: "John",
		},
		{
			name: "friends",
			want: `[{"name":"Alice","age":25},{"name":"Bob","age":28}]`,
		},
		{
			name: "friends[*].age",
			want: `~=25|28`,
		},
		{
			name: "address",
			want: `{"city":"New York","zipcode":"10001"}`,
		},
		{
			name: "json",
			want: jsonData,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.HasPrefix(tt.name, "$.") {
				tt.name = "$." + tt.name
			}
			expr, err := jp.ParseString(tt.name)
			if err != nil {
				t.Errorf("JsonChecker() error = %v", err)
				return
			}
			ck, _ := checker.Parse(tt.want)
			if got := CheckJson([]byte(jsonData), expr, ck); !got {
				t.Errorf("JsonChecker() = %v, want %v", got, tt.want)
			}
		})
	}
}
