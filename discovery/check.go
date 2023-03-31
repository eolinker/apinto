package discovery

// IHealthChecker 健康检查接口
type IHealthChecker interface {
	Check(nodes INodes)
	Reset(conf interface{}) error
	Stop()
}
