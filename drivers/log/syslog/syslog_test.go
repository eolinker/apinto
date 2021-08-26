package syslog

import (
	"github.com/eolinker/eosc/log"
	transporter_manager "github.com/eolinker/eosc/log/transporter-manager"
	syslog_transporter "github.com/eolinker/goku/log-transport/syslog"
	"testing"
)

func TestSysLog(t *testing.T) {
	c := &syslog_transporter.Config{
		Network: "",
		RAddr:   "",
		Level:   4, // info
	}
	syslog := &syslog{
		id:                 "123@log",
		name:               "TesetSysLog",
		config:             c,
		formatterName:      "line",
		transporterManager: transporter_manager.GetTransporterManager(""),
	}

	// 测试启动日志
	err := syslog.Start()
	if err != nil {
		t.Error("启动日志 失败")
		return
	}

	/* 输出内容到日志文件， 当前syslog的level是INFO,
	日志信息的level为info，warn，error，fatal，panic才能写入日志文件，
	比info的level低的debug,trace则不能写入
	level级别排序 panic,fatal,error,warn,info,debug,trace
	*/
	log.Info("syslog单元测试文件输出日志内容TEST  INFO")
	log.Warn("syslog单元测试文件输出日志内容TEST  warn")

	// 测试Reset
	newDriverConfig := &DriverConfig{
		Name:          "Tesetsyslog",
		Driver:        "syslog",
		Network:       "",
		RAddr:         "",
		Level:         "info",
		FormatterName: "json",
	}
	err = syslog.Reset(newDriverConfig, nil)
	if err != nil {
		t.Error("更新配置 失败")
		return
	}
	log.Info("syslog单元测试文件输出日志内容TEST——配置更新后  INFO")
	log.Warn("syslog单元测试文件输出日志内容TEST——配置更新后  warn")

	select {}
}
