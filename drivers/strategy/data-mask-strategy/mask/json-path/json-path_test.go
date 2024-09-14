package json_path

import (
	"testing"

	"github.com/eolinker/apinto/drivers/strategy/data-mask-strategy/mask"
)

func maskFunc(input string) string {
	return "***" // Simple mask function that replaces the input with asterisks
}

func TestJsonPath(t *testing.T) {
	tests := []struct {
		name    string
		rule    *mask.Rule
		input   string
		want    string
		wantErr bool
	}{
		{
			name: "Simple Mask",
			rule: &mask.Rule{
				Match: &mask.BasicItem{
					Type:  mask.MatchJsonPath,
					Value: "$.password",
				},
				Mask: &mask.Mask{},
			},
			input:   `{"user":"john","password":"secret123"}`,
			want:    `{"user":"john","password":"***"}`,
			wantErr: false,
		},
		{
			name: "No Match",
			rule: &mask.Rule{
				Match: &mask.BasicItem{
					Type:  mask.MatchJsonPath,
					Value: "$.notpresent",
				},
				Mask: &mask.Mask{},
			},
			input:   `{"user":"john","password":"secret123"}`,
			want:    `{"user":"john","password":"secret123"}`,
			wantErr: false,
		},
		{
			name: "Invalid JSON",
			rule: &mask.Rule{
				Match: &mask.BasicItem{
					Type:  mask.MatchJsonPath,
					Value: "$.password",
				},
				Mask: &mask.Mask{},
			},
			input:   `{"user":"john","password": "secret123"`,
			want:    "",
			wantErr: true,
		},
		{
			name: "Invalid JSONPath",
			rule: &mask.Rule{
				Match: &mask.BasicItem{
					Type:  mask.MatchJsonPath,
					Value: "$[password[",
				},
				Mask: &mask.Mask{},
			},
			input:   `{"user":"john","password":"secret123"}`,
			want:    "",
			wantErr: true,
		},
		{
			name: "Long JSON String",
			rule: &mask.Rule{
				Match: &mask.BasicItem{
					Type:  mask.MatchJsonPath,
					Value: "$.longKey",
				},
				Mask: &mask.Mask{},
			},
			input: `{"longKey":"1234567890123456789012345678901234567890"}`,
			want:  `{"longKey":"***"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.input)
			d, err := newDriver(tt.rule, maskFunc)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("newDriver() error = %v,wantErr %v", err, tt.wantErr)
					return
				}
				return
			}

			if err == nil {
				got, err := d.Exec([]byte(tt.input))
				if (err != nil) != tt.wantErr {
					t.Errorf("Exec() error = %v,wantErr %v", err, tt.wantErr)
					return
				}

				if string(got) != tt.want {
					t.Errorf("Exec() got = %v,want %v", string(got), tt.want)
				}
			}
		})
	}
}
