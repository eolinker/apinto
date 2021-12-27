package fileoutput

import "errors"

var (
	errorConfigType    = errors.New("error config type")
	errorFormatterType = errors.New("error formatter type")
	errorNilConfig = errors.New("error nil config")
	errorConfDir = errors.New("error dir is illegal")
	errorConfFile = errors.New("error file is illegal")
	errorConfPeriod = errors.New("error period is illegal")
)
