## GOKU

[![Go Report Card](https://goreportcard.com/badge/github.com/eolinker/goku)](https://goreportcard.com/report/github.com/eolinker/goku) [![Releases](https://img.shields.io/github/release/eolinker/goku/all.svg?style=flat-square)](https://github.com/eolinker/goku/releases) [![LICENSE](https://img.shields.io/github/license/eolinker/goku.svg?style=flat-square)](https://github.com/eolinker/goku/blob/main/LICENSE)

![](http://data.eolinker.com/course/ZjVKwg65f0af2f992b0ce0fcfd64d04da1696dcab3853ee.png)

------------

Goku API gateway is a microservice gateway developed based on golang. It can achieve the purposes of high-performance HTTP API forwarding, multi tenant management, API access control, etc. it has a powerful custom plug-in system, which can be expanded by itself, and can quickly help enterprises manage API services and improve the stability and security of API services.

**Note**：The **main** branch is the main development branch. Frequent updates may lead to unstable use. If you need to use a stable version, please look [release](https://github.com/eolinker/goku/releases)

### Summary / [中文介绍](https://github.com/eolinker/goku/blob/main/README_CN.md)

- [WhyGoku](#WhyGoku "Why Goku")
- [Feature](#Feature)
- [Benchmark](#Benchmark)
- [Deployment](#Deployment)
- [GetStart](#GetStart "Get Start")
- [Contact](#Contact)
- [About](#About)

### Why Goku

Goku API gateway is a microservice gateway running on the service boundary of enterprise system. When you build websites, apps, iots and even open API transactions, Goku API gateway can help you extract duplicate components from your internal system and run them on Goku gateway, such as user authorization, access control, firewall, data conversion, etc; Moreover, Goku provides the function of service arrangement, so that enterprises can quickly obtain the required data from various services and realize rapid response to business.

Goku API gateway has the following advantages:

- Completely open source: the Goku project is initiated and maintained by eolinker for a long time. We hope to work with global developers to build the infrastructure of micro service ecology.
- Excellent performance: under the same environment, Goku is about 50% faster than nginx, Kong and other products, and its stability is also optimized.
- Rich functions: Goku provides all the functions of a standard gateway, and you can quickly connect your micro services and manage network traffic.
- Extremely low use and maintenance cost: Goku is an open source gateway developed in pure go language. It has no cumbersome deployment and no external product dependence. It only needs to download and run, which is extremely simple.
- Good scalability: most of Goku's functions are modular, so you can easily expand its capabilities.

In a word, Goku API gateway enables the business development team to focus more on business implementation.

[![Stargazers over time](https://starchart.cc/eolinker/goku.svg)](#)

### Feture

| Feture            | Description                                                  |
| ----------------- | ------------------------------------------------------------ |
| Dynamic router    | Match the corresponding service by setting parameters such as location, query, header, host and method |
| Service discovery | Support such as Eureka, Nacos and Consul                     |
| Load Balance      | Support polling weight algorithm                             |
| Authentication    | Anonymous, basic, apikey, JWT, AK / SK authentication        |
| SSL certificate   | Manage multiple certificates                                 |
| Access Domain     | The access domain can be set for the gateway                 |
| Health check      | Support health check of load nodes to ensure service robustness |
| Protocol          | HTTP/HTTPS、Webservice                                       |
| Plugin            | The process is plug-in, and the required modules are loaded on demand |
| OPEN API          | Gateway configuration using open API is supported            |
| Log               | Provide the operation log of the node, and set the level output of the log |

#### RoadMap

- **Cluster**：Use **raft ** algorithm to build clusters to ensure high program availability
- **Cli**：Support cli command control gateway program
- **UI**： The gateway configuration can be operated through the UI interface, and different UI interfaces (Themes) can be customized by loading as required
- **Multi protocol**：Support a variety of protocols, including but not limited to grpc, websocket, TCP / UDP and Dubbo
- **Traffic control**：Intercept abnormal traffic
- **Black and white list**：Set the static IP black-and-white list to intercept illegal IP
- **Plugin Market**：Because Goku mainly loads the required modules through plug-in loading, users can compile the required functions into plug-ins, or download and update the plug-ins developed by contributors from the plug-in market for one click installation
- **Service Orchestration**：An orchestration API corresponds to multiple backends. The input parameters of backends support client input and parameter transfer between backends; The returned data of backend supports filtering, deleting, moving, renaming, unpacking and packaging of fields; The orchestration API can set the exception return when the orchestration call fails
- **Monitor**：Capture the gateway request data and export it to Promethus and graphite for analysis
- .....

#### RoadMap  for 2021

![image](https://user-images.githubusercontent.com/25589530/131605703-698222c6-42fb-4242-b47d-d962d949cdcf.png)

### Benchmark


![](http://data.eolinker.com/course/6Md3iDR8e64ebc99af18b628851c0b75a8a2061b4b26ff1.png)



### Deployment

* Direct Deployment：[Deployment Tutorial](https://help.gokuapi.com/?path=/quick/arrange)
* [Quick Start Tutorial](https://help.gokuapi.com/?path=/quick/quick_course)
* [Source Code Compilation Tutorial](https://help.gokuapi.com/?path=/quick/arrange)
* Docker：Follow up support
* Kubernetes：Follow up support

### Get start

1. Download and unzip the installation package (here is an example of the installation package of version v0.1.0)

```
wget https://github.com/eolinker/goku/releases/download/v0.1.0/goku-v0.1.0.linux.x64.tar.gz && tar -zxvf goku-v0.1.0.linux.x64.tar.gz && cd goku
```

2. Start gateway：

```
./goku -data_path {data_path}
```

### Contact

- **Help documentation**：[https://help.gokuapi.com](https://help.gokuapi.com)
- **QQ group**: 725853895
- **Slack**：[加入我们](https://join.slack.com/t/slack-zer6755/shared_invite/zt-u7wzqp1u-aNA0XK9Bdb3kOpN03jRmYQ)
- **Official website**：[https://www.gokuapi.com](https://www.gokuapi.com)

### About

Eolinker is a leading API management service provider, providing professional API R & D management, API automation testing, API monitoring, API gateway and other services to more than 3000 enterprises around the world. It is the first enterprise to formulate API R & D management industry specifications for ITSS (China Electronics Industry Standardization Technology Association).

Official website：[https://www.eolinker.com](https://www.eolinker.com "EOLINKER官方网站")
Download PC desktop for free：[https://www.eolinker.com/pc/](https://www.eolinker.com/pc/ "免费下载PC客户端")
