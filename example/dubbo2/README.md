# dubbo2接入Apinto

## 背景

为了方便用户快速测试apinto的dubbo2协议透传功能，本仓库提供dubbo2客户端、dubbo2服务端简易代码，使用者可按需修改使用。

## 快速使用

### 一、启动dubbo2服务端

1. 进入到**server**目录，编译dubbo2服务端程序。

```shell
cd server/ && go build -o  dubbo2Server
```
2. 启动dubbo2服务端程序
```shell
./dubbo2Server
```
### 二、配置dubbo2路由

此处使用**[Apinto-Dashboard](https://github.com/eolinker/apinto-dashboard)**进行配置演示，若为部署**Apinto-Dashboard**，可点击**[Apinto-Dashboard教程](https://help.apinto.com/docs/dashboard/quick/arrange.html)**快速部署使用。

**注意**：创建路由之前，需要确保已经新建了上游服务，若未新建，可点击**[上游服务教程](https://help.apinto.com/docs/dashboard/service/http.html#%E5%8A%9F%E8%83%BD%E6%8F%8F%E8%BF%B0)**进一步了解。
1. 进入路由列表页面，点击 "**创建**"，**Driver**选择**dubbo2**。
![](http://data.eolinker.com/course/KDIaPJ234cb011aa208cef4d98407c7e3ff630215ba1487.png)
2. 填写dubbo2配置，为了方便验证dubbo2的不同传输模式的调用情况，方法名在此示例中不填。
![](http://data.eolinker.com/course/CM1uxQH02039054d3d7a2081f5e0e4432adfb178df0935f.png)

| 字段     | 说明                                                                                                      |
|---------------------------------------------------------------------------------------------------------| ------------------------------------------------------------ |
| 端口号   | 路由监听端口号，该端口必须是**apinto**程序的config.yml中已经存在的端口号，详情请点击[程序配置说明](https://help.apinto.com/docs/apinto/quick/quick_course.html#%E8%AF%A6%E7%BB%86%E6%AD%A5%E9%AA%A4%E8%AF%B4%E6%98%8E) |
| 服务名   | dubbo2服务名                                                                                               |
| 方法名   | dubbo2方法名，不填则默认匹配该服务下的所有方法                                                                              |
| 路由规则 | 可规定客户端请求的attachment参数，                                  |
| 目标服务 | 路由匹配成功后，将转发到指定上游服务                                                                                      |
| 插件模版 | 插件模版引用                                                                                                  |
| 重试次数 | 当上游服务连接失败、连接超时时，重新转发的次数                                                                                 |
| 超时时间 | 请求上游服务的总时间                                                                                              |
至此，路由配置完成

### 三、启动dubbo2客户端

1. 进入到**client**目录，编译dubbo2客户端程序

```shell
cd client/ && go build -o dubbo2Client
```

2. 启动dubbo2客户端程序
```shell
./dubbo2Client -addr dubbo2服务端地址
```

**启动参数说明**

| 参数名     | 参数说明                          |
| ---------- |-------------------------------|
| addr       | dubbo2服务端地址，默认：127.0.0.1:8099 |

示例命令
```
./dubbo2Client -addr 127.0.0.1:8099
```
输出消息如下：
```
2023-02-17T10:42:30.641+0800    INFO   getty/getty_client.go:75        use default getty client config
2023-02-17T10:42:30.667+0800    INFO   dubbo/dubbo_protocol.go:98      [DUBBO Protocol] Refer service: dubbo://172.30.244.240:8099/api.Server?interface=api.Server&serialization=hessian2&timeout=3s
2023-02-17T10:42:30.687+0800    INFO   client/main.go:91       ComplexServer result={"addr":"192.168.0.1","server":{"age":0,"email":"1324204490@qq.com","id":16,"name":"apinto"},"time":"2023-02-17T10:42:30.641+08:00"}
2023-02-17T10:42:30.687+0800    INFO   dubbo/dubbo_protocol.go:98      [DUBBO Protocol] Refer service: dubbo://172.30.244.240:8099/api.Server?interface=api.Server&serialization=hessian2&timeout=3s
2023-02-17T10:42:30.709+0800    INFO   client/main.go:153      List result=[{"age":10,"email":"apinto1@qq.com","id":10,"name":"apinto1"},{"age":20,"email":"apinto2@qq.com","id":20,"name":"apinto2"},{"age":0,"email":"1324204
490@qq.com","id":16,"name":"apinto"}]
2023-02-17T10:42:30.709+0800    INFO   dubbo/dubbo_protocol.go:98      [DUBBO Protocol] Refer service: dubbo://172.30.244.240:8099/api.Server?interface=api.Server&serialization=hessian2&timeout=3s
2023-02-17T10:42:30.740+0800    INFO   client/main.go:174      GetById result={"age":20,"email":"apinto@qq.com","id":101,"name":"apinto"}
2023-02-17T10:42:30.740+0800    INFO   dubbo/dubbo_protocol.go:98      [DUBBO Protocol] Refer service: dubbo://172.30.244.240:8099/api.Server?interface=api.Server&serialization=hessian2&timeout=3s
2023-02-17T10:42:30.760+0800    INFO   client/main.go:113      UpdateList result=[{"age":0,"email":"1324204490@qq.com","id":16,"name":"hello"},{"age":0,"email":"1324204490@qq.com","id":16,"name":"hello"}]
2023-02-17T10:42:30.761+0800    INFO   dubbo/dubbo_protocol.go:98      [DUBBO Protocol] Refer service: dubbo://172.30.244.240:8099/api.Server?interface=api.Server&serialization=hessian2&timeout=3s
2023-02-17T10:42:30.776+0800    INFO   client/main.go:133      Update result=null

```