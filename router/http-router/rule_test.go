package http_router

import (
	"testing"

	"github.com/eolinker/apinto/router"
)

func TestRoot_Add(t *testing.T) {

	type args struct {
		id      string
		handler router.IRouterHandler
		port    int
		hosts   []string
		methods []string
		path    string
		append  []router.AppendRule
	}
	var tests = []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "1",
			args: args{
				id:      "1",
				handler: nil,
				port:    0,
				hosts:   nil,
				methods: nil,
				path:    "/",
				append:  nil,
			},
			wantErr: false,
		},
		{
			name: "2",
			args: args{
				id:      "2",
				handler: nil,
				port:    80,
				hosts:   nil,
				methods: nil,
				path:    "/",
				append:  nil,
			},
			wantErr: false,
		},
		{
			name: "3",
			args: args{
				id:      "3",
				handler: nil,
				port:    0,
				hosts:   nil,
				methods: nil,
				path:    "/",
				append:  nil,
			},
			wantErr: true,
		}, {
			name: "4",
			args: args{
				id:      "4",
				handler: nil,
				port:    0,
				hosts:   []string{"host1", "host2"},
				methods: []string{"GET", "POST"},
				path:    "/",
				append:  nil,
			},
			wantErr: false,
		},
		{
			name: "5",
			args: args{
				id:      "5",
				handler: nil,
				port:    0,
				hosts:   []string{"host1"},
				methods: []string{"GET"},
				path:    "/",
				append:  nil,
			},
			wantErr: true,
		},
	}
	r := NewRoot()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := r.Add(tt.args.id, tt.args.handler, tt.args.port, nil, tt.args.hosts, tt.args.methods, tt.args.path, tt.args.append); err != nil {
				if tt.wantErr {
					t.Logf("Add() error = %v\n", err)
				} else {
					t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}
