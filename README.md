## Apinto

[![Go Report Card](https://goreportcard.com/badge/github.com/eolinker/apinto)](https://goreportcard.com/report/github.com/eolinker/apinto) [![Releases](https://img.shields.io/github/release/eolinker/apinto/all.svg?style=flat-square)](https://github.com/eolinker/apinto/releases) [![LICENSE](https://img.shields.io/github/license/eolinker/Apinto.svg?style=flat-square)](https://github.com/eolinker/apinto/blob/main/LICENSE)![](https://shields.io/github/downloads/eolinker/apinto/total)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)](CODE_OF_CONDUCT.md)
![](http://data.eolinker.com/course/eaC48Js3400ffd03c21e36b3eea434dce22d7877a3194f6.png)

Apinto is a microservice gateway developed based on golang. It can achieve the purposes of high-performance HTTP API forwarding, multi tenant management, API access control, etc. it has a powerful user-defined plug-in system, which can be expanded by itself, and can quickly help enterprises manage API services and improve the stability and security of API services. In the future, we will provide the plug-in market. Through the strong plug-in expansion ability of **Apinto**, users can expand **Apinto** plug-ins as needed like Lego blocks to enrich **Apinto** capabilities.

**Note**：The **main** branch is the main development branch. Frequent updates may lead to unstable use. If you need to use a stable version, please look [release](https://github.com/eolinker/apinto/releases)

**Apinto** integrates configuration and forwarding functions. Users can configure it through OpenAPI or through visual UI items [apinto dashboard](https://github.com/eolinker/apinto-dashboard) for configuration, click [apinto dashboard deployment document](https://help.apinto.com/docs/dashboard/quick/arrange) for relevant documents

### Summary / [中文介绍](https://github.com/eolinker/apinto/blob/main/README_CN.md)

- [Why Apinto](#WhyApinto "Why Apinto")
- [Feature](#Feature)
- [Benchmark](#Benchmark)
- [Deployment](#Deployment)
- [GetStart](#GetStart "Get Start")
- [Contact](#Contact)
- [About](#About)

### Why Apinto

Apinto API gateway is a microservice gateway running on the service boundary of enterprise system. When you build websites, apps, iots and even open API transactions, Apinto API gateway can help you extract duplicate components from your internal system and run them on Apinto gateway, such as user authorization, access control, firewall, data conversion, etc; Moreover, Apinto provides the function of service arrangement, so that enterprises can quickly obtain the required data from various services and realize rapid response to business.

Apinto API gateway has the following advantages:

- Completely open source: the Apinto project is initiated and maintained by eolinker for a long time. We hope to work with global developers to build the infrastructure of micro service ecology.
- Excellent performance: under the same environment, Apinto is about 50% faster than nginx, Kong and other products, and its stability is also optimized.
- Rich functions: Apinto provides all the functions of a standard gateway, and you can quickly connect your micro services and manage network traffic.
- Extremely low use and maintenance cost: Apinto is an open source gateway developed in pure go language. It has no cumbersome deployment and no external product dependence. It only needs to download and run, which is extremely simple.
- Good scalability: most of Apinto's functions are modular, so you can easily expand its capabilities.

In a word, Apinto API gateway enables the business development team to focus more on business implementation.

### Star History

[![Star History Chart](https://api.star-history.com/svg?repos=eolinker/apinto&type=Date)](https://star-history.com/#eolinker/apinto&Date)


### Feture

| Feture                | Description                                                  |
| --------------------- | ------------------------------------------------------------ |
| Dynamic router        | Match the corresponding service by setting parameters such as location, query, header, host and method |
| Service discovery     | Support such as Eureka, Nacos and Consul                     |
| Load Balance          | Support polling weight algorithm                             |
| Authentication        | Anonymous, basic, apikey, JWT, AK / SK authentication        |
| SSL certificate       | Manage multiple certificates                                 |
| Access Domain         | The access domain can be set for the gateway                 |
| Health check          | Support health check of load nodes to ensure service robustness |
| Protocol              | HTTP/HTTPS、Webservice                                       |
| Plugin                | The process is plug-in, and the required modules are loaded on demand |
| OPEN API              | Gateway configuration using open API is supported            |
| Log                   | Provide the operation log of the node, and set the level output of the log |
| Multiple log output   | The node's request log can be output to different log receivers, such as file, NSQ, Kafka,etc |
| Cli                   | The gateway is controlled by cli command. The plug-in installation, download, opening and closing of the gateway can be controlled by one click command |
| Black and white list  | Support setting black-and-white list IP to intercept illegal IP |
| Parameter mapping     | Mapping the request parameters of the client to the forwarding request, you can change the location and name of the parameters as needed |
| Additional parameters | When forwarding the request, add back-end verification parameters, such as apikey, etc |
| Proxy rewrite         | It supports rewriting of 'scheme', 'URI', 'host', and adding or deleting the value of the request header of the forwarding request |
| flow control          | Intercept abnormal traffic                                   |

#### RoadMap

- **UI**： The gateway configuration can be operated through the UI interface, and different UI interfaces (Themes) can be customized by loading as required
- **Multi protocol**：Support a variety of protocols, including but not limited to grpc, websocket, TCP / UDP and Dubbo
- **Plugin Market**：Because Apinto mainly loads the required modules through plug-in loading, users can compile the required functions into plug-ins, or download and update the plug-ins developed by contributors from the plug-in market for one click installation
- **Service Orchestration**：An orchestration API corresponds to multiple backends. The input parameters of backends support client input and parameter transfer between backends; The returned data of backend supports filtering, deleting, moving, renaming, unpacking and packaging of fields; The orchestration API can set the exception return when the orchestration call fails
- **Monitor**：Capture the gateway request data and export it to Promethus and graphite for analysis
- .....

#### RoadMap  for 2022

![roadmap_en](https://user-images.githubusercontent.com/14105999/170408557-478830d5-3725-4fbe-a6f6-0ff0dd91d90e.jpeg)


### Benchmark

![image](https://user-images.githubusercontent.com/25589530/149748340-dc544f79-a8f9-46f5-903d-a3af4fb8b16e.png)



### Deployment

* Direct Deployment：[Deployment Tutorial](https://help.apinto.com/docs/apinto/quick/arrange.html)
* [Quick Start Tutorial](https://help.apinto.com/docs/apinto/quick/quick_course.html)
* [Source Code Compilation Tutorial](https://help.apinto.com/docs/apinto/quick/arrange.html)
* [Docker](https://hub.docker.com/r/eolinker/apinto-gateway)
* Kubernetes：Follow up support

### Get start

1. Download and unzip the installation package (here is an example of the installation package of version v0.9.0)

```
wget https://github.com/eolinker/apinto/releases/download/v0.8.4/apinto_v0.9.0_linux_amd64.tar.gz && tar -zxvf apinto_v0.9.0_linux_amd64.tar.gz && cd apinto
```
Apinto supports running on the arm64, i386 and amd64 architectures. 

Please download the installation package of the corresponding architecture and system as required. [Click](https://github.com/eolinker/apinto/releases/) to jump to download the installation package.

2. Start gateway：

```
./apinto start
```

3.To configure the gateway through the visual interface, click [apinto dashboard](https://github.com/eolinker/apinto-dashboard)

### Contact
- **Help documentation**：[https://help.apinto.com](https://help.apinto.com/docs)
- **QQ group**: 725853895
- **Slack**：[Join us](https://join.slack.com/t/slack-zer6755/shared_invite/zt-u7wzqp1u-aNA0XK9Bdb3kOpN03jRmYQ)
- **Official website**：[https://www.apinto.com](https://www.apinto.com)
- **Forum**：[https://community.apinto.com](https://community.apinto.com)
- **Wechat**：<img src="https://user-images.githubusercontent.com/25589530/149860447-5879437b-3cda-4833-aee3-69a2e538e85d.png" style="width:150px" />


### About

Eolink is a leading API management service provider, providing professional API R & D management, API automation testing, API monitoring, API gateway and other services to more than 3000 enterprises around the world. It is the first enterprise to formulate API R & D management industry specifications for ITSS (China Electronics Industry Standardization Technology Association).

Official website：[https://www.eolink.com](https://www.eolink.com "EOLINK官方网站")
Download PC desktop for free：[https://www.eolink.com/pc/](https://www.eolink.com/pc/ "免费下载PC客户端")
