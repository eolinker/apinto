/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package getty

import (
	"crypto/tls"
	"net"
	"reflect"

	"dubbo.apache.org/dubbo-go/v3/config"
	"dubbo.apache.org/dubbo-go/v3/protocol"
	"dubbo.apache.org/dubbo-go/v3/protocol/dubbo"
	"dubbo.apache.org/dubbo-go/v3/protocol/invocation"
	"dubbo.apache.org/dubbo-go/v3/remoting"
	gxsync "github.com/dubbogo/gost/sync"
	"github.com/eolinker/eosc/log"

	"github.com/eolinker/apinto/dubbo-getty"
)

type ServerOption func(*Server)

// Server define getty server
type Server struct {
	conf           ServerConfig
	addr           string
	codec          remoting.Codec
	tcpServer      getty.Server
	rpcHandler     *RpcServerHandler
	requestHandler func(*invocation.RPCInvocation) protocol.RPCResult
	listener       net.Listener
}

func WithAddrServer(addr string) ServerOption {
	return func(server *Server) {
		server.addr = addr
	}
}

func WithListenerServer(listener net.Listener) ServerOption {
	return func(server *Server) {
		server.listener = listener
	}
}

func WithConfigServer(conf ServerConfig) ServerOption {
	return func(server *Server) {
		server.conf = conf
	}
}

const (
	CronPeriod = 20e9
)

func init() {
	codec := &dubbo.DubboCodec{}
	remoting.RegistryCodec("dubbo", codec)
}

// NewServer create a new Server
func NewServer(handlers func(*invocation.RPCInvocation) protocol.RPCResult, serverOption ...ServerOption) *Server {
	serverConfig := GetDefaultServerConfig()
	s := &Server{
		conf:           *serverConfig,
		codec:          remoting.GetCodec("dubbo"),
		requestHandler: handlers,
	}

	for _, f := range serverOption {
		f(s)
	}

	s.rpcHandler = NewRpcServerHandler(s.conf.SessionNumber, s.conf.sessionTimeout, s)

	return s
}

func (s *Server) newSession(session getty.Session) error {
	var (
		ok bool
		//tcpConn *net.TCPConn
		//err     error
	)
	conf := s.conf

	if conf.GettySessionParam.CompressEncoding {
		session.SetCompressType(getty.CompressZip)
	}
	if _, ok = session.Conn().(*tls.Conn); ok {
		session.SetName(conf.GettySessionParam.SessionName)
		session.SetMaxMsgLen(conf.GettySessionParam.MaxMsgLen)
		session.SetPkgHandler(NewRpcServerPackageHandler(s))
		session.SetEventListener(s.rpcHandler)
		session.SetReadTimeout(conf.GettySessionParam.tcpReadTimeout)
		session.SetWriteTimeout(conf.GettySessionParam.tcpWriteTimeout)
		session.SetCronPeriod((int)(conf.heartbeatPeriod.Nanoseconds() / 1e6))
		session.SetWaitTime(conf.GettySessionParam.waitTimeout)
		log.DebugF("server accepts new session:%s\n", session.Stat())
		return nil
	}

	//session.Conn 为 *cmux.MuxConn

	log.Infof("session.Conn = %v", reflect.TypeOf(session.Conn()))

	//if _, ok = session.Conn().(*net.TCPConn); !ok {
	//	panic(fmt.Sprintf("%s, session.conn{%#v} is not tcp connection\n", session.Stat(), session.Conn()))
	//}

	if _, ok = session.Conn().(*tls.Conn); !ok {
		//if tcpConn, ok = session.Conn().(*net.TCPConn); !ok {
		//	return perrors.New(fmt.Sprintf("%s, session.conn{%#v} is not tcp connection", session.Stat(), session.Conn()))
		//}
		//
		//if err = tcpConn.SetNoDelay(conf.GettySessionParam.TcpNoDelay); err != nil {
		//	return err
		//}
		//if err = tcpConn.SetKeepAlive(conf.GettySessionParam.TcpKeepAlive); err != nil {
		//	return err
		//}
		//if conf.GettySessionParam.TcpKeepAlive {
		//	if err = tcpConn.SetKeepAlivePeriod(conf.GettySessionParam.keepAlivePeriod); err != nil {
		//		return err
		//	}
		//}
		//if err = tcpConn.SetReadBuffer(conf.GettySessionParam.TcpRBufSize); err != nil {
		//	return err
		//}
		//if err = tcpConn.SetWriteBuffer(conf.GettySessionParam.TcpWBufSize); err != nil {
		//	return err
		//}
	}

	conf.GettySessionParam.MaxMsgLen = 128 * 1024
	session.SetMaxMsgLen(conf.GettySessionParam.MaxMsgLen)
	session.SetPkgHandler(NewRpcServerPackageHandler(s))
	session.SetEventListener(s.rpcHandler)
	session.SetReadTimeout(conf.GettySessionParam.tcpReadTimeout)
	session.SetWriteTimeout(conf.GettySessionParam.tcpWriteTimeout)

	session.SetCronPeriod(CronPeriod)
	session.SetWaitTime(conf.GettySessionParam.waitTimeout)
	log.DebugF("server accepts new session: %s", session.Stat())
	return nil
}

// Start dubbo server.
func (s *Server) Start() {
	var (
		addr      string
		tcpServer getty.Server
	)
	var serverOpts []getty.ServerOption
	addr = s.addr
	if addr != "" {
		serverOpts = append(serverOpts, getty.WithLocalAddress(addr))
	}

	if s.listener != nil {
		serverOpts = append(serverOpts, getty.WithListenerServerCert(s.listener))
	}

	if s.conf.SSLEnabled {
		serverOpts = append(serverOpts, getty.WithServerSslEnabled(s.conf.SSLEnabled),
			getty.WithServerTlsConfigBuilder(config.GetServerTlsConfigBuilder()))
	}

	serverOpts = append(serverOpts, getty.WithServerTaskPool(gxsync.NewTaskPoolSimple(s.conf.GrPoolSize)))

	tcpServer = getty.NewTCPServer(serverOpts...)
	tcpServer.RunEventLoop(s.newSession)
	log.DebugF("s bind addr{%s} ok!", s.addr)
	s.tcpServer = tcpServer
}

// Stop dubbo server
func (s *Server) Stop() {
	s.tcpServer.Close()
}
