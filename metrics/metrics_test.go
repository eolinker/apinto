package metrics

import (
	"reflect"
	"testing"
)

type LabelReaderTest map[string]string

func (m LabelReaderTest) GetLabel(name string) string {
	return m[name]
}

func Test_metricsList_Metrics(t *testing.T) {
	type args struct {
		ctx LabelReader
	}
	tests := []struct {
		name string
		ms   metricsList
		args args
		want string
	}{
		{
			name: "test",
			ms:   metricsList{metricsConst("name"), metricsLabelReader("name")},
			args: args{
				ctx: LabelReaderTest{
					"name": "test",
				},
			},
			want: "name-test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ms.Metrics(tt.args.ctx); got != tt.want {
				t.Errorf("Metrics() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	ctx := LabelReaderTest{
		"name": "test",
	}
	type args struct {
		metrics []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				metrics: []string{"name", "{name}"},
			},
			want: "name-test",
		},
		{
			name: "skip1",
			args: args{
				metrics: []string{"name", "{name}", ""},
			},
			want: "name-test",
		},
		{
			name: "skip2",
			args: args{
				metrics: []string{"name", "{name}", "{}"},
			},
			want: "name-test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := Parse(tt.args.metrics)
			got := metrics.Metrics(ctx)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
