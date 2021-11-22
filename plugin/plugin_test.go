package plugin

import (
	"reflect"
	"testing"
)

func TestMergeConfig(t *testing.T) {
	type args struct {
		high map[string]*Config
		low  map[string]*Config
	}
	tests := []struct {
		name string
		args args
		want map[string]*Config
	}{
		{
			name: "high nil",
			args: args{
				high: nil,
				low: map[string]*Config{
					"low": {
						Disable: false,
						Config:  "low",
					},
				},
			},
			want: map[string]*Config{
				"low": {
					Disable: false,
					Config:  "low",
				},
			},
		}, {
			name: "low nil",
			args: args{
				high: map[string]*Config{
					"high": {Disable: false, Config: "high"},
				},
			},
			want: map[string]*Config{
				"high": {Disable: false, Config: "high"},
			},
		}, {
			name: "merge",
			args: args{
				high: map[string]*Config{
					"high": {Disable: false, Config: "high"},
				}, low: map[string]*Config{
					"low": {
						Disable: false,
						Config:  "low",
					},
				},
			},
			want: map[string]*Config{
				"high": {Disable: false, Config: "high"},
				"low": {
					Disable: false,
					Config:  "low",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeConfig(tt.args.high, tt.args.low); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
