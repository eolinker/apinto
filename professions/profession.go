/*
 * Copyright (c) 2021. Lorem ipsum dolor sit amet, consectetur adipiscing elit.
 * Morbi non lorem porttitor neque feugiat blandit. Ut vitae ipsum eget quam lacinia accumsan.
 * Etiam sed turpis ac ipsum condimentum fringilla. Maecenas magna.
 * Proin dapibus sapien vel ante. Aliquam erat volutpat. Pellentesque sagittis ligula eget metus.
 * Vestibulum commodo. Ut rhoncus gravida arcu.
 */

package professions

import (
	"github.com/eolinker/eosc"
)

func ApintoProfession() []*eosc.ProfessionConfig {
	return []*eosc.ProfessionConfig{
		{
			Name:         "router",
			Label:        "路由",
			Desc:         "路由",
			Dependencies: []string{"service"},
			AppendLabels: []string{"host", "target", "listen"},
			Drivers: []*eosc.DriverConfig{
				{
					Id:     "eolinker.com:apinto:http_router",
					Name:   "http",
					Label:  "http",
					Desc:   "http路由",
					Params: nil,
				},
			},
			Mod: eosc.ProfessionConfig_Worker,
		},

		{
			Name:         "service",
			Label:        "服务",
			Desc:         "服务",
			Dependencies: []string{"upstream"},
			AppendLabels: []string{"upstream"},
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
			Name:         "upstream",
			Label:        "上游/负载",
			Desc:         "上游/负载",
			Dependencies: []string{"discovery"},
			AppendLabels: []string{"discovery"},
			Drivers: []*eosc.DriverConfig{
				{
					Id:    "eolinker.com:apinto:upstream_http_proxy",
					Name:  "http_proxy",
					Label: "http转发负载",
					Desc:  "http转发负载",
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
					Desc:  "consul服务发现",
				},
			},
			Mod: eosc.ProfessionConfig_Worker,
		},
		{
			Name:         "auth",
			Label:        "鉴权",
			Desc:         "鉴权",
			Dependencies: nil,
			AppendLabels: nil,
			Drivers: []*eosc.DriverConfig{
				{
					Id:    "eolinker.com:apinto:auth_basic",
					Name:  "basic",
					Label: "basic鉴权",
					Desc:  "basic鉴权",
				},
				{
					Id:    "eolinker.com:apinto:auth_apikey",
					Name:  "apikey",
					Label: "apikey鉴权",
					Desc:  "apikey鉴权",
				},
				{
					Id:    "eolinker.com:apinto:auth_aksk",
					Name:  "aksk",
					Label: "ak/sk鉴权",
					Desc:  "ak/sk鉴权",
				},
				{
					Id:    "eolinker.com:apinto:auth_jwt",
					Name:  "jwt",
					Label: "jwt鉴权",
					Desc:  "jwt鉴权",
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
					Id:    "eolinker.com:goku:nsqd",
					Name:  "nsqd",
					Label: "NSQ输出",
					Desc:  "NSQ输出",
				},
				{
					Id:    "eolinker.com:goku:http_output",
					Name:  "http_output",
					Label: "http输出",
					Desc:  "http输出",
				},
				{
					Id:    "eolinker.com:goku:syslog_output",
					Name:  "syslog_output",
					Label: "syslog输出",
					Desc:  "syslog输出",
				},
				{
					Id:    "eolinker.com:goku:kafka_output",
					Name:  "kafka_output",
					Label: "kafka输出",
					Desc:  "kafka输出",
				},
			},
			Mod: eosc.ProfessionConfig_Worker,
		},
		{
			Name:         "setting",
			Label:        "setting",
			Desc:         "系统设置",
			Dependencies: nil,
			AppendLabels: nil,
			Drivers: []*eosc.DriverConfig{
				{
					Id:     "eolinker.com:apinto:plugin",
					Name:   "plugin",
					Label:  "plugin",
					Desc:   "插件管理器",
					Params: nil,
				},
			},
			Mod: eosc.ProfessionConfig_Singleton,
		},
	}
}
