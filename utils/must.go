package utils

import "encoding/json"

// MustSliceString 断言输入的参数为字符串切片
func MustSliceString(v interface{}) ([]string, error) {

	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var rs []string
	err = json.Unmarshal(data, &rs)
	if err != nil {
		return nil, err
	}
	return rs, nil
}
