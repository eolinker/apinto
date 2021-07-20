package stores

import "github.com/eolinker/eosc"

type IStoresFactory interface{
	CreateStore(config map[string]string) (eosc.IStore, error)
}

