package file_access_log

import "errors"

var (
	errorConfigType    = errors.New("error config type")
	errorFormatterType = errors.New("error formatter type")
)
