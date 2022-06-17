package nsq

import "errors"

var (
	errConfigType      = errors.New("config type does not match. ")
	errNsqConfNull     = errors.New("config is null. ")
	errTopicNull       = errors.New("topic can not be null. ")
	errAddressNull     = errors.New("Address can not be null. ")
	errFormatterType   = errors.New("type is illegal. ")
	errFormatterConf   = errors.New("formatter config can not be null. ")
	errNoValidProducer = errors.New("no valid producer. ")
	errProducerInvalid = errors.New("the producer is invalid. ")
)
