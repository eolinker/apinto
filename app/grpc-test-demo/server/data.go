package server

import (
	"encoding/json"
	"fmt"
	"io"
	"liujian-test/grpc-test-demo/common/flag"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"google.golang.org/grpc/metadata"
)

func getService(fullMethod string, name string) (interface{}, error) {
	uri := fmt.Sprintf("%s/api/service", flag.ConfigAddress)
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	type response struct {
		Data map[string]interface{} `json:"data"`
	}
	data := new(response)
	err = json.Unmarshal(body, data)
	if err != nil {
		return nil, err
	}
	path := strings.Split(fullMethod, "/")
	path = append(path, name)
	return get(data.Data["service"], path[1:])
}

func get(obj interface{}, path []string) (interface{}, error) {
	if obj == nil {
		return nil, nil
	}
	if len(path) == 0 {
		return obj, nil
	}
	currentField := path[0]
	switch om := obj.(type) {
	case map[string]interface{}:
		field := om[currentField]
		if len(path) == 1 {
			return field, nil
		}
		return get(field, path[1:])
	case *map[string]interface{}:
		return get(*om, path)
	default:
		return nil, fmt.Errorf("cannot read field:%s from type %s", currentField, reflect.TypeOf(om).String())
	}
}

func parseMetadata(value interface{}) metadata.MD {
	val, ok := value.(map[string]interface{})
	if ok {
		md := map[string]string{}
		for k, v := range val {
			switch tmp := v.(type) {
			case string:
				md[k] = tmp
			case int:
				md[k] = strconv.Itoa(tmp)
			case bool:
				md[k] = strconv.FormatBool(tmp)
			}
		}
		return metadata.New(md)
	}
	return nil
}

func retrieveData(method string, name string) (string, metadata.MD, error) {
	value, err := getService(fmt.Sprintf("/service.Hello/%s", method), name)
	if err != nil {
		return "", nil, err
	}
	data := []byte("")
	mapData, ok := value.(map[string]interface{})
	if !ok {
		data, _ = json.Marshal(value)
		return string(data), nil, nil
	}

	md := metadata.MD{}
	isMarshal := false
	for key, val := range mapData {
		switch key {
		case "data":
			data, _ = json.Marshal(val)
			isMarshal = true
		case "metadata":
			md = parseMetadata(val)
			//case "trailing-metadata":
			//	tmt = parseMetadata(val)
		}
	}
	if !isMarshal {
		data, _ = json.Marshal(mapData)
	}
	return string(data), md, nil
}
