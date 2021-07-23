# HealthCheckHTTP用法：

* 用法一：自身实现了IHealthChecker接口，可在服务发现创建app时作为healthChecker传入。
* 用法二：HTTPCheck作为健康检查中心，可生成agent，每个agent作为app的healthChecker

#### 用法一：

相当于为每个app启动一个协程，定时检查该app中所有的不健康节点

#### 用法二：

HTTPCheck作为检测中心，生成的agent作为每个app的healthChecker。当app的节点出问题时，agent充当搬运工，把节点通过通道传给HTTPCheck，交由HTTPCheck进行定时检测。相当于只为服务发现启动一个协程。


![avatar](healthcheckhttp/health_check_http.png)
