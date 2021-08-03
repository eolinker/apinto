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

### 运行

#### 运行参数说明
* -http：管理端http监听端口，http和https端口必须填写一个

* -https：管理端https监听端口，http和https端口必须填写一个

* -pem：证书文件路径，证书文件的后缀名一般为.crt 或 .pem。当需要https监听时该值必填

* -key：密钥文件路径，密钥文件的后缀名一般为.key。当需要https监听时该值必填

* -path：程序启动时加载profession路径，可填写具体的文件名，也可以填写对应的目录名

* -driver_path：驱动配置加载路径，填写具体的文件名，不填则默认为profession.yml

#### 启动程序
```
./goku-eosc -http 8081 -path data.yml
```

#### 快速使用
##### 使用openAPI配置服务
1、新建服务
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

2、新建路由
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

至此，带有路由的服务转发便已完成

3、访问服务
```
curl -X POST -H "Host: www.demo.com"\
--url http://localhost:8080/demo/params/print \
--data "username=admin&password=123456"
```

返回数据截图如下：

![](http://data.eolinker.com/course/6aNzFgWf6ed215b99024f80436866685cf8b4fe4f3b9210.png)

##### 使用文件配置服务
1、创建一个名为demo.yaml的文件，并在文件中填写下列数据
```
router:
  -
    name: demo_router
    driver: http  # http驱动
    listen: 8080    # 监听端口
    host:       # 域名列表，满足该监听端口且域名在下列列表中的可进入该路由
      - www.demo.com
    rules:      # 规则列表
      - location: "^=/demo"
    target: params_service@service
service:
  - 
	name: params_service
	driver: http
	desc: 请求参数处理
	retry: 2
	rewrite_url: /
	scheme: http
	timeout: 3000
	upstream: demoapi.gokuapi.com
```

2、启动程序
```
./goku-eosc -http 8081 -path demo.yml
```

3、访问服务
```
curl -X POST -H "Host: www.demo.com"\
--url http://localhost:8080/demo/params/print \
--data "username=admin&password=123456"
```

返回数据截图如下：

![](http://data.eolinker.com/course/6aNzFgWf6ed215b99024f80436866685cf8b4fe4f3b9210.png)