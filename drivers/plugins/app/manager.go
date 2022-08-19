package app

import "github.com/eolinker/eosc"

type Manager struct {
	AuthFactory map[string]eosc.IExtenderDriverFactory
}
