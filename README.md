# 简介
Goku 是基于 Golang 开发的开源微服务API网关

Goku 有以下的特性：
- 完全开源：Goku 项目由 Eolinker 发起并长期维护，我们希望与全球开发者共同打造微服务生态的基础设施。
- 优异的性能表现：相同环境下，Goku比Nginx、Kong等产品快约50%，并且在稳定性上也有所优化。
- 丰富的功能：Goku 提供了一个标准网关应有的所有功能，并且你可以快速连接你的各个微服务以及管理网络流量。
- 极低的使用和维护成本：Goku 是纯 Go 语言开发的开源网关，没有繁琐的部署，没有外部产品依赖，只需要下载并运行即可，极为简单。
- 良好的扩展性：Goku 的绝大部分功能都是模块化的，因此你可以很容易扩展它的能力。
- 快速与第三方工具连接：后续支持从 Swagger、Postman、Eolinker 工具导入数据，并与 Skywalking、Promethus、Graphite 等无缝连接。

# 主要功能：
|  功能 | 描述  |
| ------------ | ------------ |
|  动态路由 |  可通过设置location、query、header、host、method等参数匹配对应的服务 |
| 服务发现  | 支持对接Eureka、Nacos、Consul  |
| 负载均衡  |  支持轮询权重算法 |
|  用户鉴权 | 匿名、Basic、Apikey、JWT、AK/SK认证  |
|  SSL证书 | 管理多个证书  |
|  访问域名 | 可为网关设置访问域名  |
| 健康检查| 支持对负载的节点进行健康检查，确保服务健壮性|
|  协议 |  	HTTP/HTTPS、Webservice |
|  组件化 |  可自定义开发网关组件，并按需加载使用|
|OPEN API| 支持使用open api配置网关|

# 部署
Goku 完全基于 Golang 开发，不基于现有第三方产品，因此具有外部依赖少，部署简单等特点。

各位可以通过以下方式进行部署：

## 下载官方提供的安装包安装（推荐）
访问https://github.com/eolinker/goku/releases，
下载最新的release包，并通过以下命令安装即可：
tar -zxvf goku-v0.1.0.linux.x64.tar.gz && cd goku/（具体压缩包需要根据release提供的文件名进行修改）
./goku -http=8081 -data_path data.yaml

## 自行编译源码进行安装
访问https://github.com/eolinker/goku ，下载源码并自行编译、安装

## Docker、一键部署脚本等（后续支持）
由于目前Goku网关没有第三方依赖，整个Goku网关只有一个文件，因此不需要Docker也可以快速部署。
后续集群功能完善后会提供Docker安装包方便在windows和mac等环境下试用。


# 快速使用

## 使用步骤

1、创建服务

2、绑定路由

流程示意图如下：

![](http://data.eolinker.com/course/msPfUJBab6e929932e8610f1326d27d694a1088582b6d8c.png)

--------------
## 详细步骤说明

**Goku**支持下面两种方式进行网关配置：
* 配置文件：启动时加载初始化配置，在网关使用过程中可使用Open Api进行后续配置，支持使用Open Api导出配置

* openAPI：可在网关使用过程中动态配置网关信息，包括路由、服务、负载均衡、鉴权、服务发现等

路由配置规则详情点此[查看](http://help.gokuapi.com/?path=/router_driver/http)

### 使用文件配置服务


#### 创建一个名为demo.yaml的文件，并在文件中填写下列数据

```
router:
  -
    name: demo_router
    driver: http  # http驱动
    listen: 8080    # 监听端口
    host:       # 域名列表，满足该监听端口且域名在下列列表中的可进入该路由
      - www.demo.com
    rules:      # 规则列表
      - location: "/demo*"		# 匹配带有"/demo"前缀的url
    target: params_service@service	# 绑定服务ID，服务ID为：{name}@service，其中name为服务的名称
service:
  - 
	name: params_service	# 服务名称
	driver: http			# 驱动名称
	desc: 请求参数处理	# 服务描述
	retry: 2			# 重试次数
	rewrite_url: /		# 转发重写url
	scheme: http		# 请求上游协议
	timeout: 3000	  # 超时时间，单位ms
	upstream: www.gokuapi.com	# 上游地址，该处可填写负载ID或上游的域名/IP+端口
```

#### 启动程序

```
./goku -http 8081 -data_path demo.yml
```

#### 访问服务

```
curl -X POST -H "Host: www.demo.com"\
--url http://localhost:8080/demo/params/print \
--data "username=admin&password=123456"
```

返回数据截图如下：

![](http://data.eolinker.com/course/6aNzFgWf6ed215b99024f80436866685cf8b4fe4f3b9210.png)

### 使用openAPI配置网关

在程序启动后，我们可以通过openAPI进行动态配置网关信息，包括路由、服务、鉴权、负载均衡、服务发现等

#### 创建服务

```
curl -i -X POST \
--url http://localhost:8081/api/service \
-H "Content-Type: application/json" \
--data "{
    "name": "params_service",
    "driver": "http",
    "desc": "请求参数处理",
    "timeout": 3000,
    "upstream": "demoapi.gokuapi.com",
    "retry": 3,
    "rewrite_url": "/",
    "scheme": "https"
}"
```

请求参数说明如下：

![](http://data.eolinker.com/course/9hDUGZz764dddfd79f3aca4bd1b7f284d61dcf15ee1735b.png)

返回数据说明如下：

![](http://data.eolinker.com/course/6faVYZ1b781c2e7fe13d4f5e6a6893de68e559c806c3c6f.png)

返回数据示例：
```
{
    "id": "params_service@service",
    "name": "params_service",
    "driver": "http",
    "create_time": "2021-08-03 14:31:50",
    "update_time": "2021-08-03 14:31:50"
}
```

#### 创建路由，并且服务id绑定路由

将第1步返回的 **id** 值填入到路由配置的 **target** 中，如上例中的 **id** 为 **params_service@service**

```
curl -i -X POST \
--url http://localhost:8081/api/router \
-H "Content-Type: application/json" \
--data "{
    "name": "demo_router",
    "driver": "http",
    "desc": "绑定8080端口",
    "listen": 8080,
    "host": ["www.demo.com"],
    "rules": [{
        "location": "/demo"
    }],
    "target": "params_service@service"
}"
```

请求参数说明如下：

![](http://data.eolinker.com/course/Qf5BWg459891c9ab414a94a447bde0059cf327f70c6232e.png)

返回数据说明如下：

![](http://data.eolinker.com/course/6faVYZ1b781c2e7fe13d4f5e6a6893de68e559c806c3c6f.png)

返回数据示例：
```
{
    "id": "demo_router@service",
    "name": "demo_router",
    "driver": "http",
    "create_time": "2021-08-03 14:31:50",
    "update_time": "2021-08-03 14:31:50"
}
```

至此，带有路由的服务转发配置完成

#### 访问服务

```
curl -X POST -H "Host: www.demo.com"\
--url http://localhost:8080/demo/params/print \
--data "username=admin&password=123456"
```

返回数据截图如下：

![](http://data.eolinker.com/course/6aNzFgWf6ed215b99024f80436866685cf8b4fe4f3b9210.png)

