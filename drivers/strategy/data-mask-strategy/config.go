package data_mask_strategy

import (
	"github.com/eolinker/apinto/drivers/strategy/data-mask-strategy/mask"
	"github.com/eolinker/apinto/strategy"
)

type Config struct {
	Name        string `json:"name" skip:"skip"`
	Description string `json:"description" skip:"skip"`
	//Stop        bool                  `json:"stop"`
	Priority int                   `json:"priority" label:"优先级" description:"1-999"`
	Filters  strategy.FilterConfig `json:"filters" label:"过滤规则"`
	DataMask mask.DataMask         `json:"data_mask" label:"规则"`
}
