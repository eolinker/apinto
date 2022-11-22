//go:build !windows
// +build !windows

/*
 * Copyright (c) 2021. Lorem ipsum dolor sit amet, consectetur adipiscing elit.
 * Morbi non lorem porttitor neque feugiat blandit. Ut vitae ipsum eget quam lacinia accumsan.
 * Etiam sed turpis ac ipsum condimentum fringilla. Maecenas magna.
 * Proin dapibus sapien vel ante. Aliquam erat volutpat. Pellentesque sagittis ligula eget metus.
 * Vestibulum commodo. Ut rhoncus gravida arcu.
 */

package main

import (
	"os"

	"github.com/eolinker/apinto/utils/version"

	"github.com/eolinker/eosc/env"

	"github.com/eolinker/eosc"

	"github.com/eolinker/eosc/eoscli"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/eosc/process"
)

func init() {

	process.Register(eosc.ProcessWorker, ProcessWorker)
	process.Register(eosc.ProcessMaster, ProcessMaster)
	process.Register(eosc.ProcessHelper, ProcessHelper)
	process.Register(eosc.ProcessAdmin, ProcessAdmin)

	InitCLILog()
}

func main() {

	if process.Run() {
		log.Close()
		return
	}
	if env.IsDebug() {
		if process.RunDebug(eosc.ProcessMaster) {
			log.Info("debug done")
		} else {
			log.Warn("debug not exist")
		}
		log.Close()
		return
	}
	app := eoscli.NewApp()
	app.Default()
	app.AppendCommand(version.Build())
	err := app.Run(os.Args)
	if err != nil {
		log.Error(err)
	}
	log.Close()
}
