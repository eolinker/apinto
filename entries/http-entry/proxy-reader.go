package http_entry

import (
	"strconv"

	http_service "github.com/eolinker/eosc/eocontext/http-context"
)

type IProxyReader interface {
	ReadProxy(name string, proxy http_service.IProxy) (interface{}, bool)
}

func ReadProxyFromProxyReader(reader IProxyReader, proxy http_service.IProxy, key string) (string, bool) {
	var data string
	value, has := reader.ReadProxy(key, proxy)
	if !has {
		return "", false
	}
	switch v := value.(type) {
	case string:
		data = v
	case []byte:
		data = string(v)
	case int:
		data = strconv.Itoa(v)
	case int64:
		data = strconv.FormatInt(v, 10)
	case float32:
		data = strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		data = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		data = strconv.FormatBool(v)
	default:
		return "", false
	}
	return data, true
}

type ProxyReadFunc func(name string, proxy http_service.IProxy) (interface{}, bool)

func (p ProxyReadFunc) ReadProxy(name string, proxy http_service.IProxy) (interface{}, bool) {
	return p(name, proxy)
}
