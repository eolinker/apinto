package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eolinker/apinto/resources"
	"github.com/eolinker/eosc/eocontext"
	http_service "github.com/eolinker/eosc/eocontext/http-context"
	"net/http"
	"time"
)

type ResponseData struct {
	Header    http.Header
	Body      []byte
	ValidTime int
	Now       time.Time // 缓存存放的时间
}

func (r *ResponseData) Complete(ctx eocontext.EoContext) error {
	httpCtx, err := http_service.Assert(ctx)
	if err != nil {
		return err
	}
	httpCtx.Response().SetBody(r.Body)
	for key, val := range r.Header {
		if len(val) > 0 {
			httpCtx.Response().SetHeader(key, val[0])
		}
	}
	httpCtx.Response().SetHeader("Date", time.Now().Format(time.RFC822))

	//计算Age  Age 的值通常接近于 0。表示此对象刚刚从原始服务器获取不久；其他的值则是表示代理服务器当前的系统时间与此应答中的通用头 Date 的值之差
	age := int(time.Now().Sub(r.Now).Seconds())

	httpCtx.Response().Headers().Set("Age", fmt.Sprintf("%d", age))
	httpCtx.Response().Headers().Set("Cache-Control", fmt.Sprintf("%s=%d", "max-age", r.ValidTime))

	return nil
}

func SetResponseData(cache resources.ICache, uri string, data *ResponseData, validTime int) {
	bytes, _ := json.Marshal(data)
	cache.Set(context.TODO(), uri, bytes, time.Second*time.Duration(validTime))
}

func GetResponseData(cache resources.ICache, uri string) *ResponseData {
	result := cache.Get(context.TODO(), uri)
	bytes, _ := result.Bytes()
	data := new(ResponseData)
	if err := json.Unmarshal(bytes, data); err != nil {
		return nil
	}
	return data
}
