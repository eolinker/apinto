package cache

import (
	"encoding/json"
	"github.com/coocood/freecache"
)

var freeCache *freecache.Cache

func NewCache() {
	freeCache = freecache.NewCache(0)
}

type ResponseData struct {
	Header map[string]string
	Body   []byte
}

func SetCache(uri string, data *ResponseData, validTime int) {
	bytes, _ := json.Marshal(data)
	_ = freeCache.Set([]byte(uri), bytes, validTime)
}

func GetCache(uri string) *ResponseData {
	bytes, _ := freeCache.Get([]byte(uri))
	data := new(ResponseData)
	if err := json.Unmarshal(bytes, data); err != nil {
		return nil
	}
	return data
}
