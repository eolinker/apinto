package cache

type ICache interface {
	Set(key string, value)
}
