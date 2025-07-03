package router

import (
	"errors"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/soheilhy/cmux"
)

type RouterType int

const (
	GRPC RouterType = iota
	Http
	Dubbo2
	TlsTCP
	AnyTCP
	depth
)

var (
	handlers = make([]ServerHandler, depth)
	//matchers                 = make([][]cmux.Matcher, depth)
	matchWriters             = make([][]cmux.MatchWriter, depth)
	ErrorDuplicateRouterType = errors.New("duplicate")
)

func Register(tp RouterType, handler ServerHandler) error {
	if handlers[tp] != nil {
		return ErrorDuplicateRouterType
	}
	handlers[tp] = handler
	return nil
}

type ServerHandler func(port int, listener net.Listener)

func readPort(addr net.Addr) int {
	ipPort := addr.String()
	i := strings.LastIndex(ipPort, ":")
	port := ipPort[i+1:]
	pv, _ := strconv.Atoi(port)
	return pv
}

func matchersToMatchWriters(matchers ...cmux.Matcher) []cmux.MatchWriter {
	mws := make([]cmux.MatchWriter, 0, len(matchers))
	for _, m := range matchers {
		cm := m
		mws = append(mws, func(w io.Writer, r io.Reader) bool {
			return cm(r)
		})
	}
	return mws
}
