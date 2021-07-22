package access_field

import (
	"encoding/json"
	"testing"
)

func TestGenSelectFieldList(t *testing.T) {
	type args struct {
		selectFields []string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "all",
			args: args{
				selectFields: nil,
			},
		}, {
			name: "test",
			args: args{
				selectFields: []string{"api_title"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenSelectFieldList(tt.args.selectFields)
			data, _ := json.MarshalIndent(got, "", "  ")
			t.Logf("GenSelectFieldList() %s = %s ", tt.name, string(data))

		})
	}
}
