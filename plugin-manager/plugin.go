package plugin_manager

const (
	StatusDisable = "disable"
	StatusEnable  = "enable"
	StatusGlobal  = "global"
)

type Plugins []*Plugin

type Plugin struct {
	Name   string      `json:"name"`
	ID     string      `json:"id"`
	Type   string      `json:"type"`
	Status string      `json:"status"`
	Config interface{} `json:"config"`
}
