package filelog

import (
	"github.com/eolinker/eosc/log"
	transporter_manager "github.com/eolinker/eosc/log/transporter-manager"
	filelog_transporter "github.com/eolinker/goku/log/common/filelog-transporter"
	"testing"
	"time"
)

func TestFileLog(t *testing.T) {
	period, _ := filelog_transporter.ParsePeriod("day")
	c := &filelog_transporter.Config{
		Dir:    "",
		File:   "test_filelog",
		Expire: 1,
		Period: period,
		Level:  4, // info
	}
	fileLog := &filelog{
		id:                 "123@log",
		name:               "TesetFileLog",
		config:             c,
		formatterName:      "line",
		transporterManager: transporter_manager.GetTransporterManager(""),
	}

	// 测试启动日志
	err := fileLog.Start()
	if err != nil {
		t.Error("启动日志 失败")
		return
	}

	/* 输出内容到日志文件， 当前fileLog的level是INFO,
	日志信息的level为info，warn，error，fatal，panic才能写入日志文件，
	比info的level低的debug,trace则不能写入
	level级别排序 panic,fatal,error,warn,info,debug,trace
	*/
	log.Info("filelog单元测试文件输出日志内容TEST  INFO")
	log.Warn("filelog单元测试文件输出日志内容TEST  warn")

	// 测试Reset
	newDriverConfig := &DriverConfig{
		Name:          "TesetFileLog",
		Driver:        "filelog",
		Dir:           "",
		File:          "test_filelog2",
		Level:         "info",
		Period:        "day",
		Expire:        1,
		FormatterName: "json",
	}
	err = fileLog.Reset(newDriverConfig, nil)
	if err != nil {
		t.Error("更新配置 失败")
		return
	}
	log.Info("filelog单元测试文件输出日志内容TEST——配置更新后  INFO")
	log.Warn("filelog单元测试文件输出日志内容TEST——配置更新后  warn")
	time.Sleep(1 * time.Second)
}
