package router

const (
	group      = "goku"
	version    = "v1.0"
	label      = "http路由"
	desc       = "http路由"
	name       = "http"
	profession = "router"
)

type Config struct {
	ID     string   `json:"id"`
	Name   string   `json:"name" yaml:"name"`
	Driver string   `json:"driver" yaml:"driver"`
	Listen string   `json:"listen" yaml:"listen"`
	Host   []string `json:"host" yaml:"host"`
	Rules  []Rule   `json:"rules" yaml:"rules"`
}

type Rule struct {
	Location string            `json:"location" yaml:"location"`
	Header   map[string]string `json:"header" yaml:"header"`
	Query    map[string]string `json:"query" yaml:"query"`
	Target   string            `json:"target" target:"target"`
}
