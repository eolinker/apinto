package main

import (
	"github.com/eolinker/eosc/log"
	"os"
)

func InitCLILog() {
	formatter := &log.LineFormatter{
		TimestampFormat:  "2006-01-02 15:04:05",
		CallerPrettyfier: nil,
	}
	transport := log.NewTransport(os.Stdout, log.ErrorLevel)
	transport.SetFormatter(formatter)
	log.Reset(transport)
}
