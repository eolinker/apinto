package acl

import (
	"github.com/eolinker/apinto/drivers"
	"github.com/eolinker/eosc"
)

type Config struct {
	Allow            []string `json:"allow"`
	Deny             []string `json:"deny"`
	HideGroupsHeader bool     `json:"hide_groups_header"`
}

func Create(id, name string, conf *Config, workers map[eosc.RequireId]eosc.IWorker) (eosc.IWorker, error) {
	allow := make(map[string]struct{})
	deny := make(map[string]struct{})
	for _, a := range conf.Allow {
		allow[a] = struct{}{}
	}
	for _, d := range conf.Deny {
		deny[d] = struct{}{}
	}
	return &executor{
		WorkerBase: drivers.Worker(id, name),
		cfg:        conf,
		allow:      allow,
		deny:       deny,
	}, nil
}
