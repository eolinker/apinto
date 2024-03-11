package manager

import "github.com/eolinker/apinto/router"

func Set(id string, port int, hosts []string, method []string, path string, append []AppendRule, router router.IRouterHandler) error {
	return routerManager.Set(id, port, nil, hosts, method, path, append, router)
}
func Delete(id string) {
	routerManager.Delete(id)
}
func AddPreRouter(id string, method []string, path string, handler router.IRouterPreHandler) {
	routerManager.AddPreRouter(id, method, path, handler)
}
func DeletePreRouter(id string) {
	routerManager.DeletePreRouter(id)
}
