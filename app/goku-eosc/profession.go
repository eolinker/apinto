/*
 * Copyright (c) 2021. Lorem ipsum dolor sit amet, consectetur adipiscing elit.
 * Morbi non lorem porttitor neque feugiat blandit. Ut vitae ipsum eget quam lacinia accumsan.
 * Etiam sed turpis ac ipsum condimentum fringilla. Maecenas magna.
 * Proin dapibus sapien vel ante. Aliquam erat volutpat. Pellentesque sagittis ligula eget metus.
 * Vestibulum commodo. Ut rhoncus gravida arcu.
 */

package main

import "github.com/eolinker/eosc"

func professionConfig() []eosc.ProfessionConfig {
	pcs := []eosc.ProfessionConfig{
		{
			Name:         "router",
			Label:        "路由",
			Desc:         "路由",
			Dependencies: []string{"service"},
			AppendLabel:  []string{"host", "service"},

			Drivers: []eosc.DriverConfig{
				{
					ID:     "eolinker:goku:http_router",
					Name:   "http",
					Label:  "http",
					Desc:   "http路由",
					Params: nil,
				},
			},
		}, {
			Name:         "service",
			Label:        "服务",
			Desc:         "服务",
			Dependencies: []string{"upstream"},
			AppendLabel:  []string{"upstream"},
			Drivers: []eosc.DriverConfig{
				{
					ID:     "eolinker:goku:service_http",
					Name:   "http",
					Label:  "service",
					Desc:   "服务",
					Params: nil,
				},
			},
		},
		{
			Name:         "upstream",
			Label:        "上游/负载",
			Desc:         "上游/负载",
			Dependencies: []string{"discovery"},
			AppendLabel:  []string{"discovery"},
			Drivers: []eosc.DriverConfig{
				{
					ID:     "eolinker:goku:http_proxy",
					Name:   "http_proxy",
					Label:  "http转发负载",
					Desc:   "http转发负载",
					Params: nil,
				},
			},
		}, {
			Name:         "discovery",
			Label:        "注册中心",
			Desc:         "注册中心",
			Dependencies: []string{},
			AppendLabel:  []string{},
			Drivers: []eosc.DriverConfig{
				{
					ID:     "eolinker:goku:discovery_static",
					Name:   "static",
					Label:  "静态服务发现",
					Desc:   "静态服务发现",
					Params: nil,
				},
				{
					ID:     "eolinker:goku:discovery_nacos",
					Name:   "nacos",
					Label:  "nacos服务发现",
					Desc:   "nacos服务发现",
					Params: nil,
				},
				{
					ID:     "eolinker:goku:discovery_consul",
					Name:   "consul",
					Label:  "consul服务发现",
					Desc:   "consul服务发现",
					Params: nil,
				},
				{
					ID:     "eolinker:goku:discovery_eureka",
					Name:   "eureka",
					Label:  "eureka服务发现",
					Desc:   "eureka服务发现",
					Params: nil,
				},
			},
		},
	}
	return pcs
}
