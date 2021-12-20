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

func GokuProfession() []*eosc.ProfessionConfig {
	return []*eosc.ProfessionConfig{
		{
			Name:         "router",
			Label:        "路由",
			Desc:         "路由",
			Dependencies: []string{"service"},
			AppendLabels: []string{"host", "target", "listen"},
			Drivers: []*eosc.DriverConfig{
				{
					Id:     "eolinker.com:goku:http_router",
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
					Id:    "eolinker.com:goku:service_http",
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
					Id:    "eolinker.com:goku:upstream_http_proxy",
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
					Id:    "eolinker.com:goku:discovery_static",
					Name:  "static",
					Label: "静态服务发现",
					Desc:  "静态服务发现",
				}, {
					Id:    "eolinker.com:goku:discovery_nacos",
					Name:  "nacos",
					Label: "nacos服务发现",
					Desc:  "nacos服务发现",
				}, {
					Id:    "eolinker.com:goku:discovery_consul",
					Name:  "consul",
					Label: "consul服务发现",
					Desc:  "consul服务发现",
				}, {
					Id:    "eolinker.com:goku:discovery_eureka",
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
					Id:    "eolinker.com:goku:auth_basic",
					Name:  "basic",
					Label: "basic鉴权",
					Desc:  "basic鉴权",
				},
				{
					Id:    "eolinker.com:goku:auth_apikey",
					Name:  "apikey",
					Label: "apikey鉴权",
					Desc:  "apikey鉴权",
				},
				{
					Id:    "eolinker.com:goku:auth_aksk",
					Name:  "aksk",
					Label: "ak/sk鉴权",
					Desc:  "ak/sk鉴权",
				},
				{
					Id:    "eolinker.com:goku:auth_jwt",
					Name:  "jwt",
					Label: "jwt鉴权",
					Desc:  "jwt鉴权",
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
					Id:     "eolinker.com:goku:plugin",
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
