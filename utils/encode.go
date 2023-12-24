package utils

import (
	"encoding/base64"
	"net/url"
	"strings"
)

// B64Decode base64解密
func B64DecodeString(input string) (string, error) {
	data, err := B64Decode(input)
	if err != nil {
		return "", err
	}
	return string(data), err
}
func B64Decode(input string) ([]byte, error) {
	remainder := len(input) % 4
	// base64编码需要为4的倍数，如果不是4的倍数，则填充"="号
	if remainder > 0 {
		padlen := 4 - remainder
		input = input + strings.Repeat("=", padlen)
	}
	// 将原字符串中的"_","-"分别用"/"和"+"替换
	input = strings.Replace(strings.Replace(input, "_", "/", -1), "-", "+", -1)
	result, err := base64.StdEncoding.DecodeString(input)
	return result, err
}

// B64Encode base64加密
func B64Encode(input []byte) string {
	//result := base64.StdEncoding.EncodeToString([]byte(input))
	//base64.RawStdEncoding.EncodeToString([]byte(input))
	//result = strings.Replace(strings.Replace(strings.Replace(result, "=", "", -1), "/", "_", -1), "+", "-", -1)
	return base64.StdEncoding.EncodeToString(input)
}

// QueryUrlEncode 对query进行url encode
func QueryUrlEncode(rawQuery string) string {
	queryList := strings.Split(rawQuery, "&")
	for i, query := range queryList {
		idx := strings.Index(query, "=")
		if idx != -1 {
			queryList[i] = query[:idx] + "=" + url.QueryEscape(query[idx+1:])
		}
	}
	return strings.Join(queryList, "&")
}
