package main

import (
	ai_key "github.com/eolinker/apinto/drivers/ai-key"
	ai_provider "github.com/eolinker/apinto/drivers/ai-provider"
	"github.com/eolinker/apinto/drivers/certs"
	"github.com/eolinker/apinto/drivers/discovery/consul"
	"github.com/eolinker/apinto/drivers/discovery/eureka"
	"github.com/eolinker/apinto/drivers/discovery/nacos"
	"github.com/eolinker/apinto/drivers/discovery/static"
	"github.com/eolinker/apinto/drivers/output/loki"
	auth_interceptor "github.com/eolinker/apinto/drivers/plugins/auth-interceptor"
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
	auth_interceptor.Register(extenderRegister)

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
	loki.Register(extenderRegister)

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

	//// AI供应商
	//openAI.Register(extenderRegister)
	//
	//google.Register(extenderRegister)
	//moonshot.Register(extenderRegister)
	//tongyi.Register(extenderRegister)
	//zhipuai.Register(extenderRegister)
	//fireworks.Register(extenderRegister)
	//novita.Register(extenderRegister)
	//mistralai.Register(extenderRegister)
	//
	//anthropic.Register(extenderRegister)
	//
	//baichuan.Register(extenderRegister)
	//stepfun.Register(extenderRegister)
	//wenxin.Register(extenderRegister)
	//yi.Register(extenderRegister)
	//perfxcloud.Register(extenderRegister)
	//cohere.Register(extenderRegister)
	//deepseek.Register(extenderRegister)
	//hunyuan.Register(extenderRegister)
	//openrouter.Register(extenderRegister)
	//groq.Register(extenderRegister)
	//upstage.Register(extenderRegister)
	//minimax.Register(extenderRegister)
	//chatglm.Register(extenderRegister)
	//bedrock.Register(extenderRegister)
	//spark.Register(extenderRegister)
	//nvidia.Register(extenderRegister)
	//vertex_ai.Register(extenderRegister)
	//fakegpt.Register(extenderRegister)
	//zhinao.Register(extenderRegister)
	//ollama.Register(extenderRegister)

	ai_provider.Register(extenderRegister)
	ai_key.Register(extenderRegister)
}
