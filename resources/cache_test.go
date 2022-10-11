package resources

import "testing"

func TestReplace(t *testing.T) {
	type args struct {
		caches []ICache
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "nil",
			args: args{nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Replace(tt.args.caches...)
		})
	}
}
