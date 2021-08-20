package httplog

import (
	"github.com/eolinker/eosc/log"
	transporter_manager "github.com/eolinker/eosc/log/transporter-manager"
	httplog_transporter "github.com/eolinker/goku/log/common/httplog-transporter"
	"net/http"
	"testing"
)

func TestHTTPLog(t *testing.T) {
	c := &httplog_transporter.Config{
		Method:  "POST",
		Url:     "http://127.0.0.1:8080/test",
		Headers: http.Header{"a": []string{"1"}},
		Level:   4, // info
	}
	fileLog := &httplog{
		id:                 "123@log",
		name:               "TesetHttpLog",
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
	log.Info("httplog单元测试文件输出日志内容TEST  INFO")
	log.Warn("httplog单元测试文件输出日志内容TEST  warn")

	// 测试Reset
	newDriverConfig := &DriverConfig{
		Name:          "TesetFileLog",
		Driver:        "httplog",
		Method:        "GET",
		Url:           "http://127.0.0.1:8081/test",
		Headers:       map[string]string{},
		Level:         "info",
		FormatterName: "json",
	}
	err = fileLog.Reset(newDriverConfig, nil)
	if err != nil {
		t.Error("更新配置 失败")
		return
	}
	log.Info("httplog单元测试文件输出日志内容TEST——配置更新后  INFO")
	log.Warn("httplog单元测试文件输出日志内容TEST——配置更新后  warn")


	select{}
}
