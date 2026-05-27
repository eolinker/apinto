package access_hierarchy

import "github.com/eolinker/apinto/utils/response"

type Rule struct {
	A string `yaml:"a" json:"a,omitempty" label:"App Key" description:"App key metrics syntax" require:"true"`
	B string `yaml:"b" json:"b,omitempty" label:"Resource Key" description:"Resource key metrics syntax" require:"true"`
}

type Config struct {
	Rules    []*Rule            `yaml:"rules" json:"rules" label:"Rules"`
	Response *response.Response `yaml:"response" json:"response" label:"Response"`
}
