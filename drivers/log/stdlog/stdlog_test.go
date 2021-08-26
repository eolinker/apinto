package stdlog

import (
	"github.com/eolinker/eosc/log"
	transporter_manager "github.com/eolinker/eosc/log/transporter-manager"
	stdlog_transporter "github.com/eolinker/goku/log-transport/stdlog"
	"testing"
)

func TestStdLog(t *testing.T) {
	c := &stdlog_transporter.Config{
		Level: 4, // info
	}
	stdlog := &stdlog{
		id:                 "123@log",
		name:               "Tesetstdlog",
		config:             c,
		formatterName:      "line",
		transporterManager: transporter_manager.GetTransporterManager(""),
	}

	// 测试启动日志
	err := stdlog.Start()
	if err != nil {
		t.Error("启动日志 失败")
		return
	}

	/* 输出内容到日志文件， 当前stdlog的level是INFO,
	日志信息的level为info，warn，error，fatal，panic才能写入日志文件，
	比info的level低的debug,trace则不能写入
	level级别排序 panic,fatal,error,warn,info,debug,trace
	*/
	log.Info("stdlog单元测试文件输出日志内容TEST  INFO")
	log.Warn("stdlog单元测试文件输出日志内容TEST  warn")

	// 测试Reset
	newDriverConfig := &DriverConfig{
		Name:          "Tesetstdlog",
		Driver:        "stdlog",
		Level:         "warn",
		FormatterName: "json",
	}
	err = stdlog.Reset(newDriverConfig, nil)
	if err != nil {
		t.Error("更新配置 失败")
		return
	}
	log.Info("stdlog单元测试文件输出日志内容TEST——配置更新后  INFO")
	log.Warn("stdlog单元测试文件输出日志内容TEST——配置更新后  warn")

	select {}
}
