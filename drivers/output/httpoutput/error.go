package httpoutput

import "errors"

var (
	errConfigType    = errors.New("config type does not match. ")
	errMethod        = errors.New("method is illegal. ")
	errUrlNull       = errors.New("url can not be null. ")
	errFormatterType = errors.New("type is illegal. ")
	errFormatterConf = errors.New("formatter config can not be null. ")
)
