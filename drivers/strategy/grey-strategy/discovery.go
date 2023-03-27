package grey_strategy

import "github.com/eolinker/apinto/drivers/discovery/static"

var (
	defaultHttpDiscovery = static.CreateAnonymous(&static.Config{
		Health:   nil,
		HealthOn: false,
	})
)
