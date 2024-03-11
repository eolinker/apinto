package httpoutput

import (
	"encoding/json"
	"net/http"

	"github.com/eolinker/eosc"
)

type Config struct {
	Scopes        []string             `json:"scopes" label:"作用域"`
	Method        string               `json:"method" yaml:"method" enum:"GET,POST,PUT" label:"请求方式"`
	Url           string               `json:"url" yaml:"url" format:"uri" label:"请求Url"`
	Headers       map[string]string    `json:"headers" yaml:"headers" label:"请求头部"`
	Type          string               `json:"type" yaml:"type" enum:"json,line" label:"输出格式"`
	ContentResize []ContentResize      `json:"content_resize" yaml:"content_resize" label:"内容截断配置" switch:"type===json"`
	Formatter     eosc.FormatterConfig `json:"formatter" yaml:"formatter" label:"格式化配置"`
}

type ContentResize struct {
	Size   int    `json:"size" label:"内容截断大小" description:"单位：M" default:"10" minimum:"0"`
	Suffix string `json:"suffix" label:"匹配标签后缀"`
}

func (h *Config) isConfUpdate(conf *Config) bool {
	if h.Method != conf.Method || h.Url != conf.Url || !compareTwoMapStringEqual(h.Headers, conf.Headers) || !compareArray(h.Scopes, conf.Scopes) || h.Type != conf.Type {
		return true
	}
	return false
}

func compareTwoMapStringEqual(mapA, mapB map[string]string) bool {
	if len(mapA) != len(mapB) {
		return false
	}
	length := len(mapA)
	keySlice := make([]string, 0, length)
	dataValueA := make([]string, 0, length)
	dataValueB := make([]string, 0, length)
	for k, v := range mapA {
		keySlice = append(keySlice, k)
		dataValueA = append(dataValueA, v)
	}

	for _, key := range keySlice {
		if vb, has := mapB[key]; has {
			dataValueB = append(dataValueB, vb)
		} else {
			return false
		}
	}

	strValueA, _ := json.Marshal(dataValueA)
	strValueB, _ := json.Marshal(dataValueB)

	return string(strValueA) == string(strValueB)
}

func toHeader(items map[string]string) http.Header {
	header := make(http.Header)
	for k, v := range items {
		header.Set(k, v)
	}
	return header
}

func compareArray[T comparable](o, t []T) bool {
	if len(o) != len(t) {
		return false
	}
	oMap := make(map[T]struct{})
	tMap := make(map[T]struct{})
	for i := 0; i < len(o); i++ {
		oMap[o[i]] = struct{}{}
		tMap[t[i]] = struct{}{}
	}
	if len(oMap) != len(tMap) {
		return false
	}
	for k := range oMap {
		if _, has := tMap[k]; !has {
			return false
		}
	}
	return true
}
