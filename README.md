## GOKU

[![Go Report Card](https://goreportcard.com/badge/github.com/eolinker/goku)](https://goreportcard.com/report/github.com/eolinker/goku) [![Releases](https://img.shields.io/github/release/eolinker/goku/all.svg?style=flat-square)](https://github.com/eolinker/goku/releases) [![LICENSE](https://img.shields.io/github/license/eolinker/goku.svg?style=flat-square)](https://github.com/eolinker/goku/blob/main/LICENSE)

![](http://data.eolinker.com/course/ZjVKwg65f0af2f992b0ce0fcfd64d04da1696dcab3853ee.png)

------------

Goku API Gateway （中文名：悟空 API 网关）是一个基于 Golang 开发的微服务网关，能够实现高性能 HTTP API 转发、多租户管理、API 访问权限控制等目的，拥有强大的自定义插件系统可以自行扩展，能够快速帮助企业进行 API 服务治理、提高 API 服务的稳定性和安全性。

注意：**main**分支为开发主要分支，频繁更新可能导致使用不稳定，若需要使用稳定版本，请查看[release](https://github.com/eolinker/goku/releases)

### 概况

- [为什么要使用Goku](#为什么要使用Goku "为什么要使用Goku")
- [产品特性](#产品特性 "产品特性")
- [基准测试](#基准测试 "基准测试")
- [部署](#部署 "部署")
- [启动](#启动 "启动")
- [联系我们](#联系我们 "联系我们")
- [关于我们](#关于我们 "关于我们")
- [授权协议](#授权协议 "授权协议")

### 为什么要使用Goku

Goku API Gateway （悟空 API 网关）是运行在企业系统服务边界上的微服务网关。当您构建网站、App、IOT甚至是开放API交易时，Goku API Gateway 能够帮你将内部系统中重复的组件抽取出来并放置在Goku网关上运行，如进行用户授权、访问控制、防火墙、数据转换等；并且Goku 提供服务编排的功能，让企业可以快速从各类服务上获取需要的数据，对业务实现快速响应。

Goku API Gateway具有以下优势：

- 完全开源：Goku 项目由 Eolinker 发起并长期维护，我们希望与全球开发者共同打造微服务生态的基础设施。
- 优异的性能表现：相同环境下，Goku比Nginx、Kong等产品快约50%，并且在稳定性上也有所优化。
- 丰富的功能：Goku 提供了一个标准网关应有的所有功能，并且你可以快速连接你的各个微服务以及管理网络流量。
- 极低的使用和维护成本：Goku 是纯 Go 语言开发的开源网关，没有繁琐的部署，没有外部产品依赖，只需要下载并运行即可，极为简单。
- 良好的扩展性：Goku 的绝大部分功能都是模块化的，因此你可以很容易扩展它的能力。

总而言之，Goku API Gateway 能让业务开发团队更加专注地实现业务。

[![Stargazers over time](https://starchart.cc/eolinker/goku.svg)](#)

### 产品特性

| 功能     | 描述                                                         |
| -------- | ------------------------------------------------------------ |
| 动态路由 | 可通过设置location、query、header、host、method等参数匹配对应的服务 |
| 服务发现 | 支持对接Eureka、Nacos、Consul                                |
| 负载均衡 | 支持轮询权重算法                                             |
| 用户鉴权 | 匿名、Basic、Apikey、JWT、AK/SK认证                          |
| SSL证书  | 管理多个证书                                                 |
| 访问域名 | 可为网关设置访问域名                                         |
| 健康检查 | 支持对负载的节点进行健康检查，确保服务健壮性                 |
| 协议     | HTTP/HTTPS、Webservice                                       |
| 插件化   | 流程插件化，按需加载所需模块                                 |
| OPEN API | 支持使用open api配置网关                                     |
| 日志     | 提供节点的运行日志,可设置日志的等级输出                      |

#### 迭代计划

- **集群支持**：使用**raft**算法构建集群，保证程序高可用性

- **Cli命令支持**：不同参数值不同转发

- **UI界面支持**： 通过UI界面操作网关配置，可以通过需要加载定制不同的UI界面（主题）

- **多协议支持**：支持多种协议，包括但不限于：gRPC、Websocket、tcp/udp、Dubbo

- **流量控制**：拦截异常流量

- **黑白名单**：设置静态IP黑白名单，拦截非法IP

- **插件市场**：由于goku主要是通过插件加载的方式加载所需模块，用户可将所需功能编译成插件，也可从插件市场中下载更新贡献者开发的插件，一键安装使用

- **服务编排**：一个编排API对应多个backend，backend的入参支持客户端传入，也支持backend间的参数传递；backend的返回数据支持字段的过滤、删除、移动、重命名、拆包和封包；编排API能够设定编排调用失败时的异常返回

- **监控**：捕获网关请求数据，并可将其导出到promethus、Graphite中进行分析
- .....

#### 2021年迭代计划

![](http://data.eolinker.com/course/tbDpymJ8343df96713b8bb44b053c2088536ad59d7483d3.png)

### 基准测试


![](http://data.eolinker.com/course/6Md3iDR8e64ebc99af18b628851c0b75a8a2061b4b26ff1.png)



### 部署

* 直接部署：[部署教程](https://help.gokuapi.com/?path=/quick/arrange)
* [快速入门教程](https://help.gokuapi.com/?path=/quick/quick_course)
* [源码编译教程](https://help.gokuapi.com/?path=/quick/arrange)
* Docker部署：后续支持
* Kubernetes部署：后续支持

### 启动

1.下载安装包并解压（此处以v0.1.0版本的安装包示例）

```
wget https://github.com/eolinker/goku/releases/download/v0.1.0/goku-v0.1.0.linux.x64.tar.gz && tar -zxvf goku-v0.1.0.linux.x64.tar.gz && cd goku
```

2.启动网关：

```
./goku -data_path {配置文件路径}
```

### 联系我们

- **帮助文档**：[https://help.gokuapi.com](https://help.gokuapi.com)
- **QQ群**: 725853895
- **Slack**：[加入我们](https://join.slack.com/t/slack-zer6755/shared_invite/zt-u7wzqp1u-aNA0XK9Bdb3kOpN03jRmYQ)
- **官网**：[https://www.gokuapi.com](https://www.gokuapi.com)

### 关于我们

EOLINKER 是领先的 API 管理服务供应商，为全球超过3000家企业提供专业的 API 研发管理、API自动化测试、API监控、API网关等服务。是首家为ITSS（中国电子工业标准化技术协会）制定API研发管理行业规范的企业。

官方网站：[https://www.eolinker.com](https://www.eolinker.com "EOLINKER官方网站")
免费下载PC桌面端：[https://www.eolinker.com/pc/](https://www.eolinker.com/pc/ "免费下载PC客户端")
