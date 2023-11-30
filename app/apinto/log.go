package main

import (
	"os"

	"github.com/eolinker/eosc/env"
	"github.com/eolinker/eosc/log"
)

func InitCLILog() {
	formatter := &log.LineFormatter{
		TimestampFormat:  "2006-01-02 15:04:05",
		CallerPrettyfier: nil,
	}
	transport := log.NewTransport(os.Stdout, env.ErrorLevel())
	transport.SetFormatter(formatter)
	log.Reset(transport)
}
