package app_response_rewrite

import "github.com/eolinker/apinto/utils/response"

type Config struct {
	Response *response.Response `json:"response" label:"响应内容"`
}
