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
	level, err := log.ParseLevel(env.ErrorLevel())
	if err != nil {
		level = log.InfoLevel
	}

	transport := log.NewTransport(os.Stdout, level)
	transport.SetFormatter(formatter)
	log.Reset(transport)
}
