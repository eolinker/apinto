package main

import (
	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/log"
)

var (
	storeMemory = "memory"
)

func initStore() eosc.IStore {
	memoryDriver, has := eosc.GetStoreDriver(storeMemory)
	if !has {
		log.Panic("unkonw store driver:", storeMemory)
	}
	memoryStore, err := memoryDriver.Create(nil)
	if err != nil {
		log.Panic("memory store create error:", err)
	}
	err = memoryStore.Initialization()
	if err != nil {
		log.Panic(err)
	}

	return memoryStore
}
