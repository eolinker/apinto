package cache

import (
	"encoding/json"
	"github.com/coocood/freecache"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

var freeCache *freecache.Cache

func NewCache() {
	freeCache = freecache.NewCache(0)
}

type ResponseData struct {
	Header map[string]string
	Body   []byte
}

func (r *ResponseData) Complete(ctx eocontext.EoContext) error {
	httpCtx, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	httpCtx.Response().SetBody(r.Body)
	for key, val := range r.Header {
		httpCtx.Response().SetHeader(key, val)
	}
	return nil
}

func SetResponseData(uri string, data *ResponseData, validTime int) {
	bytes, _ := json.Marshal(data)
	_ = freeCache.Set([]byte(uri), bytes, validTime)
}

func GetResponseData(uri string) *ResponseData {
	bytes, _ := freeCache.Get([]byte(uri))
	data := new(ResponseData)
	if err := json.Unmarshal(bytes, data); err != nil {
		return nil
	}
	return data
}
