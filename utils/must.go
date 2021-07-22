package utils

import "encoding/json"

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
