### 简介
**goku-eosc**是一个基于 Golang 开发的微服务网关，其基于框架[eosc](https://github.com/eolinker/goku-eosc "eosc")进行开发

### 特性
* 灵活的路由：可通过设置**location**、**query**、**header**、**host**、**method**等参数匹配对应的服务

* 支持对接多个主流的服务发现应用
	* eureka
	
	* nacos
	
	* consul

* 支持多种鉴权
	* ak/sk
	
	* basic
	
	* apikey
	
	* jwt
* 支持健康检查

