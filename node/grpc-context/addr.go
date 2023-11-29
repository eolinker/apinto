package grpc_context

import (
	"net"
)

var zeroTCPAddr = &net.TCPAddr{
	IP: net.IPv4zero,
}

func addrToIP(addr net.Addr) net.IP {
	x, ok := addr.(*net.TCPAddr)
	if !ok {
		return net.IPv4zero
	}
	return x.IP
}
