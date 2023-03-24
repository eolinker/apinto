package discovery

// IHealthChecker 健康检查接口
type IHealthChecker interface {
	check(nodes []INode)
}
