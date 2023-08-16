/*
 * Copyright (c) 2021. Lorem ipsum dolor sit amet, consectetur adipiscing elit.
 * Morbi non lorem porttitor neque feugiat blandit. Ut vitae ipsum eget quam lacinia accumsan.
 * Etiam sed turpis ac ipsum condimentum fringilla. Maecenas magna.
 * Proin dapibus sapien vel ante. Aliquam erat volutpat. Pellentesque sagittis ligula eget metus.
 * Vestibulum commodo. Ut rhoncus gravida arcu.
 */

package main

import (
	"github.com/eolinker/apinto/drivers/certs"
	http_counter "github.com/eolinker/apinto/drivers/counter/http"
	"github.com/eolinker/apinto/drivers/discovery/consul"
	"github.com/eolinker/apinto/drivers/discovery/eureka"
	"github.com/eolinker/apinto/drivers/discovery/nacos"
	"github.com/eolinker/apinto/drivers/discovery/static"

	"github.com/eolinker/apinto/application/auth"
	"github.com/eolinker/apinto/drivers/discovery/polaris"
	"github.com/eolinker/apinto/drivers/output/fileoutput"
	"github.com/eolinker/apinto/drivers/output/httpoutput"
	"github.com/eolinker/apinto/drivers/output/kafka"
	"github.com/eolinker/apinto/drivers/output/nsq"
	"github.com/eolinker/apinto/drivers/output/syslog"
	plugin_manager "github.com/eolinker/apinto/drivers/plugin-manager"
	"github.com/eolinker/apinto/drivers/resources/datasource/influxdbv2"
	"github.com/eolinker/apinto/drivers/resources/redis"
	dubbo2_router "github.com/eolinker/apinto/drivers/router/dubbo2-router"
	grpc_router "github.com/eolinker/apinto/drivers/router/grpc-router"
	http_router "github.com/eolinker/apinto/drivers/router/http-router"
	"github.com/eolinker/apinto/drivers/service"
	cache_strategy "github.com/eolinker/apinto/drivers/strategy/cache-strategy"
	fuse_strategy "github.com/eolinker/apinto/drivers/strategy/fuse-strategy"
	grey_strategy "github.com/eolinker/apinto/drivers/strategy/grey-strategy"
	limiting_strategy "github.com/eolinker/apinto/drivers/strategy/limiting-strategy"
	visit_strategy "github.com/eolinker/apinto/drivers/strategy/visit-strategy"
	"github.com/eolinker/apinto/drivers/template"
	protocbuf "github.com/eolinker/apinto/drivers/transcode/protobuf"

	"github.com/eolinker/apinto/drivers/app"
	"github.com/eolinker/apinto/drivers/output/prometheus"
	"github.com/eolinker/eosc"
)

func ApintoProfession() []*eosc.ProfessionConfig {
	return []*eosc.ProfessionConfig{
		{
			Name:         "router",
			Label:        "路由",
			Desc:         "路由",
			Dependencies: []string{"service", "template"},
			AppendLabels: []string{"host", "service", "listen", "disable"},
			Drivers: []*eosc.DriverConfig{
				{
					Id:     "eolinker.com:apinto:http_router",
					Name:   "http",
					Label:  "http",
					Desc:   "http路由",
					Params: nil,
				},
				{
					Id:     "eolinker.com:apinto:grpc_router",
					Name:   "grpc",
					Label:  "grpc",
					Desc:   "grpc路由",
					Params: nil,
				},
				{
					Id:     "eolinker.com:apinto:dubbo2_router",
					Name:   "dubbo2",
					Label:  "dubbo2",
					Desc:   "dubbo2路由",
					Params: nil,
				},
			},
			Mod: eosc.ProfessionConfig_Worker,
		},
		{
			Name:         "service",
			Label:        "服务",
			Desc:         "服务",
			Dependencies: []string{"discovery"},
			AppendLabels: []string{"discovery"},
			Drivers: []*eosc.DriverConfig{
				{
					Id:    "eolinker.com:apinto:service_http",
					Name:  "http",
					Label: "service",
					Desc:  "服务",
				},
			},
			Mod: eosc.ProfessionConfig_Worker,
		},
		{
			Name:         "template",
			Label:        "模版",
			Desc:         "模版",
			Dependencies: nil,
			AppendLabels: nil,
			Drivers: []*eosc.DriverConfig{
				{
					Id:    "eolinker.com:apinto:plugin_template",
					Name:  "plugin_template",
					Label: "插件模版",
					Desc:  "插件模版",
				},
			},
			Mod: eosc.ProfessionConfig_Worker,
		},
		{
			Name:         "discovery",
			Label:        "注册中心",
			Desc:         "注册中心",
			Dependencies: nil,
			AppendLabels: nil,
			Drivers: []*eosc.DriverConfig{
				{
					Id:    "eolinker.com:apinto:discovery_static",
					Name:  "static",
					Label: "静态服务发现",
					Desc:  "静态服务发现",
				}, {
					Id:    "eolinker.com:apinto:discovery_nacos",
					Name:  "nacos",
					Label: "nacos服务发现",
					Desc:  "nacos服务发现",
				}, {
					Id:    "eolinker.com:apinto:discovery_consul",
					Name:  "consul",
					Label: "consul服务发现",
					Desc:  "consul服务发现",
				}, {
					Id:    "eolinker.com:apinto:discovery_eureka",
					Name:  "eureka",
					Label: "eureka服务发现",
					Desc:  "eureka服务发现",
				}, {
					Id:    "eolinker.com:apinto:discovery_polaris",
					Name:  "polaris",
					Label: "北极星服务发现",
					Desc:  "北极星服务发现",
				},
			},
			Mod: eosc.ProfessionConfig_Worker,
		},
		{
			Name:         "app",
			Label:        "应用",
			Desc:         "应用",
			Dependencies: nil,
			AppendLabels: []string{"disable"},
			Drivers: []*eosc.DriverConfig{
				{
					Id:    "eolinker.com:apinto:app",
					Name:  "app",
					Label: "应用",
					Desc:  "应用",
				},
			},
			Mod: eosc.ProfessionConfig_Worker,
		}, {
			Name:         "strategy",
			Label:        "策略",
			Desc:         "策略",
			Dependencies: nil,
			AppendLabels: nil,
			Drivers: []*eosc.DriverConfig{
				{
					Id:    "eolinker.com:apinto:strategy-limiting",
					Name:  "limiting",
					Label: "限流策略",
					Desc:  "限流策略",
				},
				{
					Id:    "eolinker.com:apinto:strategy-cache",
					Name:  "cache",
					Label: "缓存策略",
					Desc:  "缓存策略",
				},
				{
					Id:    "eolinker.com:apinto:strategy-grey",
					Name:  "grey",
					Label: "灰度策略",
					Desc:  "灰度策略",
				},
				{
					Id:    "eolinker.com:apinto:strategy-visit",
					Name:  "visit",
					Label: "访问策略",
					Desc:  "访问策略",
				},
				{
					Id:    "eolinker.com:apinto:strategy-fuse",
					Name:  "fuse",
					Label: "熔断策略",
					Desc:  "熔断策略",
				},
			},
			Mod: eosc.ProfessionConfig_Worker,
		},
		{
			Name:         "counter",
			Label:        "计数器",
			Desc:         "计数器",
			Dependencies: nil,
			AppendLabels: nil,
			Drivers: []*eosc.DriverConfig{
				{
					Id:    "eolinker.com:apinto:http-counter",
					Name:  "http",
					Label: "http计数器",
					Desc:  "http计数器",
				},
			},
			Mod: eosc.ProfessionConfig_Worker,
		},
		{
			Name:         "output",
			Label:        "输出",
			Desc:         "输出",
			Dependencies: nil,
			AppendLabels: nil,
			Drivers: []*eosc.DriverConfig{
				{
					Id:    "eolinker.com:apinto:file_output",
					Name:  "file",
					Label: "文件输出",
					Desc:  "文件输出",
				},
				{
					Id:    "eolinker.com:apinto:nsqd",
					Name:  "nsqd",
					Label: "NSQ输出",
					Desc:  "NSQ输出",
				},
				{
					Id:    "eolinker.com:apinto:http_output",
					Name:  "http_output",
					Label: "http输出",
					Desc:  "http输出",
				},
				{
					Id:    "eolinker.com:apinto:syslog_output",
					Name:  "syslog_output",
					Label: "syslog输出",
					Desc:  "syslog输出",
				},
				{
					Id:    "eolinker.com:apinto:kafka_output",
					Name:  "kafka_output",
					Label: "kafka输出",
					Desc:  "kafka输出",
				},
				{
					Id:    "eolinker.com:apinto:redis",
					Name:  "redis",
					Label: "redis 集群",
					Desc:  "redis 集群",
				},
				{
					Id:    "eolinker.com:apinto:influxdbv2",
					Name:  "influxdbv2",
					Label: "influxdbv2输出",
					Desc:  "influxdbv2输出",
				},
				{
					Id:    "eolinker.com:apinto:prometheus_output",
					Name:  "prometheus",
					Label: "prometheus输出",
					Desc:  "prometheus输出",
				},
			},
			Mod: eosc.ProfessionConfig_Worker,
		},

		{
			Name:         "certificate",
			Label:        "证书",
			Desc:         "证书",
			Dependencies: nil,
			AppendLabels: nil,
			Drivers: []*eosc.DriverConfig{
				{
					Id:    "eolinker.com:apinto:ssl-server",
					Name:  "server",
					Label: "证书",
					Desc:  "证书",
				},
			},
			Mod: eosc.ProfessionConfig_Worker,
		},
		{
			Name:         "transcode",
			Label:        "编码器",
			Desc:         "编码器",
			Dependencies: nil,
			AppendLabels: nil,
			Drivers: []*eosc.DriverConfig{
				{
					Id:    "eolinker.com:apinto:protobuf_transcode",
					Name:  "protobuf",
					Label: "protobuf编码器",
					Desc:  "protobuf编码器",
				},
			},
			Mod: eosc.ProfessionConfig_Worker,
		},
	}
}

func driverRegister(extenderRegister eosc.IExtenderDriverRegister) {
	// 路由
	http_router.Register(extenderRegister)
	grpc_router.Register(extenderRegister)
	dubbo2_router.Register(extenderRegister)

	// 上游服务
	service.Register(extenderRegister)

	// 服务发现
	static.Register(extenderRegister)
	nacos.Register(extenderRegister)
	consul.Register(extenderRegister)
	eureka.Register(extenderRegister)
	polaris.Register(extenderRegister)

	// 应用
	app.Register(extenderRegister)
	auth.Register(extenderRegister)

	// 插件相关
	plugin_manager.Register(extenderRegister)
	template.Register(extenderRegister)

	// 输出器
	fileoutput.Register(extenderRegister)
	nsq.Register(extenderRegister)
	httpoutput.Register(extenderRegister)
	kafka.Register(extenderRegister)
	syslog.Register(extenderRegister)
	prometheus.Register(extenderRegister)
	redis.Register(extenderRegister)
	influxdbv2.Register(extenderRegister)

	// 服务治理-策略
	limiting_strategy.Register(extenderRegister)
	cache_strategy.Register(extenderRegister)
	grey_strategy.Register(extenderRegister)
	visit_strategy.Register(extenderRegister)
	fuse_strategy.Register(extenderRegister)

	// 编码器
	protocbuf.Register(extenderRegister)

	// 计数器
	http_counter.Register(extenderRegister)

	// 证书
	certs.Register(extenderRegister)
}
