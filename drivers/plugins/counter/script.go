package counter

import (
	_ "embed"

	"github.com/go-redis/redis/v8"
)

var (
	//go:embed decr.lua
	lockScriptByte []byte
	//go:embed callback.lua
	callbackScriptByte []byte

	lockScript     *redis.Script
	callbackScript *redis.Script
)

func init() {
	lockScript = redis.NewScript(string(lockScriptByte))
	callbackScript = redis.NewScript(string(callbackScriptByte))
}
