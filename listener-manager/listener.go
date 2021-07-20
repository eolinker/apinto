package listener_manager

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

const (
	TCP = "tcp"
	UDP = "udp"
)

var defaultListener = newListener()

type listener struct {
	tcpListeners map[int]net.Listener
	udpListeners map[int]net.Listener
	locker       sync.RWMutex
}

func newListener() *listener {
	return &listener{tcpListeners: make(map[int]net.Listener), udpListeners: make(map[int]net.Listener), locker: sync.RWMutex{}}
}

func (l *listener) getTCPListener(port int) (net.Listener, error) {
	l.locker.RLock()
	defer l.locker.RUnlock()
	if v, ok := l.tcpListeners[port]; ok {
		return v, nil
	}
	return nil, errors.New("no listener in port")
}

func (l *listener) setTCPListener(port int, listen net.Listener) {
	l.locker.Lock()
	defer l.locker.Unlock()
	l.tcpListeners[port] = listen
}

func (l *listener) deleteTCPListener(port int) {
	l.locker.Lock()
	defer l.locker.Unlock()
	delete(l.tcpListeners, port)
}

func NewTCPListener(ip string, port int) (net.Listener, error) {
	addr := fmt.Sprintf("%s:%d", ip, port)
	l, err := net.Listen(TCP, addr)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func GetTCPListener(port int) (net.Listener, error) {
	return defaultListener.getTCPListener(port)
}

func SetTCPListener(port int, listen net.Listener) {
	defaultListener.setTCPListener(port, listen)
}

func DeleteTCPListener(port int) {
	defaultListener.deleteTCPListener(port)
}
