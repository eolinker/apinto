package router

var _ iServer = (*Server)(nil)
var _ iServers = (*Servers)(nil)

type iServer interface {
}
type Server struct {
}
type iServers interface {
}
type Servers struct {
}
