## Apinto

[![Go Report Card](https://goreportcard.com/badge/github.com/eolinker/apinto)](https://goreportcard.com/report/github.com/eolinker/apinto) [![Releases](https://img.shields.io/github/release/eolinker/apinto/all.svg?style=flat-square)](https://github.com/eolinker/apinto/releases) [![LICENSE](https://img.shields.io/github/license/eolinker/apinto.svg?style=flat-square)](https://github.com/eolinker/apinto/blob/main/LICENSE)![](https://shields.io/github/downloads/eolinker/apinto/total)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)](CODE_OF_CONDUCT_CN.md)

------------
![](http://data.eolinker.com/course/eaC48Js3400ffd03c21e36b3eea434dce22d7877a3194f6.png)

Apinto是一个基于 Golang 开发的微服务网关，能够实现高性能 HTTP API 转发、多租户管理、API 访问权限控制等目的，拥有强大的自定义插件系统可以自行扩展，能够快速帮助企业进行 API 服务治理、提高 API 服务的稳定性和安全性。未来我们将提供插件市场，通过**Apinto**强大的插件拓展能力，用户可像乐高积木一样根据需要自行拓展**Apinto**的插件，丰富**Apinto**的能力。

注意：**main**分支为开发主要分支，频繁更新可能导致使用不稳定，若需要使用稳定版本，请查看[release](https://github.com/eolinker/apinto/releases)

**Apinto** 集合了配置和转发功能，使用者可以通过openAPI进行配置，也可通过可视化UI项目[Apinto Dashboard](https://github.com/eolinker/apinto-dashboard)进行配置，相关文档可点击[Apinto Dashboard部署文档](https://help.apinto.com/docs/dashboard/quick/arrange)
### 概况 | [English Introduction](https://github.com/eolinker/apinto/blob/main/README.md)

- [为什么要使用Apinto](#为什么要使用Apinto "Apinto")
- [产品特性](#产品特性 "产品特性")
- [基准测试](#基准测试 "基准测试")
- [部署](#部署 "部署")
- [启动](#启动 "启动")
- [联系我们](#联系我们 "联系我们")
- [关于我们](#关于我们 "关于我们")
- [授权协议](#授权协议 "授权协议")

### 为什么要使用Apinto

**Apinto**是运行在企业系统服务边界上的API网关。当您构建网站、App、IOT甚至是开放API交易时，Apinto 能够帮你将内部系统中重复的组件抽取出来并放置在Apinto网关上运行，如进行用户授权、访问控制、防火墙、数据转换等；并且Apinto 提供服务编排的功能，让企业可以快速从各类服务上获取需要的数据，对业务实现快速响应。

**Apinto**具有以下优势：

- 完全开源：Apinto 项目由 Eolinker 发起并长期维护，我们希望与全球开发者共同打造微服务生态的基础设施。
- 优异的性能表现：相同环境下，Apinto比Nginx、Kong等产品快约50%，并且在稳定性上也有所优化。
- 丰富的功能：Apinto 提供了一个标准网关应有的所有功能，并且你可以快速连接你的各个微服务以及管理网络流量。
- 极低的使用和维护成本：Apinto 是纯 Go 语言开发的开源网关，没有繁琐的部署，没有外部产品依赖，只需要下载并运行即可，极为简单。
- 良好的扩展性：Apinto 的绝大部分功能都是模块化的，因此你可以很容易扩展它的能力。

总而言之，Apinto 能让业务开发团队更加专注地实现业务。

### Star 历史

[![Star History Chart](https://api.star-history.com/svg?repos=eolinker/apinto&type=Date)](https://star-history.com/#eolinker/apinto&Date)


### 产品特性

| 功能         | 描述                                                         |
| ------------ | ------------------------------------------------------------ |
| 动态路由     | 可通过设置location、query、header、host、method等参数匹配对应的服务 |
| 服务发现     | 支持对接Eureka、Nacos、Consul                                |
| 负载均衡     | 支持轮询权重算法                                             |
| 用户鉴权     | 匿名、Basic、Apikey、JWT、AK/SK认证                          |
| SSL证书      | 管理多个证书                                                 |
| 访问域名     | 可为网关设置访问域名                                         |
| 健康检查     | 支持对负载的节点进行健康检查，确保服务健壮性                 |
| 协议         | HTTP/HTTPS、Webservice、Restful                              |
| 插件化       | 流程插件化，按需加载所需模块                                 |
| OPEN API     | 支持使用open api配置网关                                     |
| 日志         | 提供节点的运行日志，可根据日志设置的等级输出                 |
| 多种日志输出 | 可将节点的请求日志输出到不同的日志接收器，如file、nsq、kafka等 |
| Cli命令支持  | 通过Cli命令操控网关，插件安装、下载和网关的开启、关闭等操作均可使用一键命令操控 |
| 黑白名单     | 支持设置黑白名单IP，拦截非法IP                               |
| 参数映射     | 将客户端的请求参数映射到转发请求中，可按需改变参数的位置及名称 |
| 额外参数     | 转发请求时，额外加上后端验证参数，如apikey等                 |
| 转发重写     | 支持对 `scheme`、`uri`、`host` 的重写，同时支持对转发请求的请求头部header的值进行新增或者删除 |
| 流量控制     | 拦截异常流量                                                 |

#### 迭代计划

- **UI界面支持**： 通过UI界面操作网关配置，可以通过需要加载定制不同的UI界面（主题）

- **多协议支持**：支持多种协议，包括但不限于：gRPC、Websocket、tcp/udp、Dubbo

- **插件市场**：由于apinto主要是通过插件加载的方式加载所需模块，用户可将所需功能编译成插件，也可从插件市场中下载更新贡献者开发的插件，一键安装使用

- **服务编排**：一个编排API对应多个backend，backend的入参支持客户端传入，也支持backend间的参数传递；backend的返回数据支持字段的过滤、删除、移动、重命名、拆包和封包；编排API能够设定编排调用失败时的异常返回

- **监控**：捕获网关请求数据，并可将其导出到promethus、Graphite中进行分析
- .....

#### 2022年迭代计划
![roadmap_cn](https://user-images.githubusercontent.com/14105999/170409057-407055ef-2d30-4272-ae8c-3c46b95af8d1.jpeg)

### 基准测试


![image](https://user-images.githubusercontent.com/25589530/149748340-dc544f79-a8f9-46f5-903d-a3af4fb8b16e.png)

### 部署

* 直接部署：[部署教程](https://help.apinto.com/docs/apinto/quick/arrange)
* [快速入门教程](https://help.apinto.com/docs/apinto/quick/quick_course)
* [源码编译教程](https://help.apinto.com/docs/apinto/quick/arrange)
* [Docker部署](https://hub.docker.com/r/eolinker/apinto-gateway)
* Kubernetes部署：后续支持

### 启动

1.下载安装包并解压（此处以v0.9.0版本的安装包示例）

```
wget https://github.com/eolinker/apinto/releases/download/v0.9.0/apinto_v0.9.0_linux_amd64.tar.gz && tar -zxvf apinto_v0.9.0_linux_amd64.tar.gz && cd apinto
```

Apinto支持在arm64、i386、amd64架构上运行。

请根据需要下载对应架构及系统的安装包，安装包下载请[点击](https://github.com/eolinker/apinto/releases/)跳转

2.启动网关：

```
./apinto start
```

3.如需可视化界面请点击[Apinto Dashboard](https://github.com/eolinker/apinto-dashboard)


- ### **联系我们**


* **帮助文档**：[https://help.apinto.com](https://help.apinto.com/docs)

- **QQ群**: 725853895

- **Slack**：[加入我们](https://join.slack.com/t/slack-zer6755/shared_invite/zt-u7wzqp1u-aNA0XK9Bdb3kOpN03jRmYQ)

- **官网**：[https://www.apinto.com](https://www.apinto.com/)
- **论坛**：[https://community.apinto.com](https://community.apinto.com/)
- **微信群**：<img src="https://user-images.githubusercontent.com/25589530/149860447-5879437b-3cda-4833-aee3-69a2e538e85d.png" style="width:150px" />

### 关于我们

EOLINK 是领先的 API 管理服务供应商，为全球超过3000家企业提供专业的 API 研发管理、API自动化测试、API监控、API网关等服务。是首家为ITSS（中国电子工业标准化技术协会）制定API研发管理行业规范的企业。

官方网站：[https://www.eolink.com](https://www.eolink.com "EOLINK官方网站")

免费下载PC桌面端：[https://www.eolink.com/pc/](https://www.eolink.com/pc/ "免费下载PC客户端")
