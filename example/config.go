package example

import "github.com/eolinker/eosc"

type Config struct {
	Name string `json:"name"`
	Label string `json:"label"`
	Target eosc.RequireId `json:"target"`
}
