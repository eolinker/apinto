package cache

//todo 本地缓存demo

type Cache struct {
	Header map[string]string
	Body   []byte
}

func SetCache(uri string, cache *Cache, validTime int) {

}

func GetCache(uri string) *Cache {
	return nil
}
