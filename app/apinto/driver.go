package main

import (
	anthropic "github.com/eolinker/apinto/drivers/ai-provider/authropic"
	"github.com/eolinker/apinto/drivers/ai-provider/baichuan"
	"github.com/eolinker/apinto/drivers/ai-provider/fireworks"
	"github.com/eolinker/apinto/drivers/ai-provider/google"
	"github.com/eolinker/apinto/drivers/ai-provider/mistralai"
	"github.com/eolinker/apinto/drivers/ai-provider/moonshot"
	"github.com/eolinker/apinto/drivers/ai-provider/novita"
	"github.com/eolinker/apinto/drivers/ai-provider/openAI"
	"github.com/eolinker/apinto/drivers/ai-provider/perfxcloud"
	"github.com/eolinker/apinto/drivers/ai-provider/stepfun"
	"github.com/eolinker/apinto/drivers/ai-provider/tongyi"
	"github.com/eolinker/apinto/drivers/ai-provider/yi"
	"github.com/eolinker/apinto/drivers/ai-provider/zhipuai"
	"github.com/eolinker/apinto/drivers/certs"
	"github.com/eolinker/apinto/drivers/discovery/consul"
	"github.com/eolinker/apinto/drivers/discovery/eureka"
	"github.com/eolinker/apinto/drivers/discovery/nacos"
	"github.com/eolinker/apinto/drivers/discovery/static"
	data_mask_strategy "github.com/eolinker/apinto/drivers/strategy/data-mask-strategy"

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
	data_mask_strategy.Register(extenderRegister)

	// 编码器
	protocbuf.Register(extenderRegister)

	// 证书
	certs.Register(extenderRegister)

	// AI供应商
	openAI.Register(extenderRegister)

	google.Register(extenderRegister)
	moonshot.Register(extenderRegister)
	tongyi.Register(extenderRegister)
	zhipuai.Register(extenderRegister)
	fireworks.Register(extenderRegister)
	novita.Register(extenderRegister)
	mistralai.Register(extenderRegister)

	anthropic.Register(extenderRegister)

	baichuan.Register(extenderRegister)
	stepfun.Register(extenderRegister)
	yi.Register(extenderRegister)
	perfxcloud.Register(extenderRegister)
}
