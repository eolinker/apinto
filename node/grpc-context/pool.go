package grpc_context

import (
	"context"
	"crypto/tls"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/keepalive"
)

var (
	ErrConnShutdown = errors.New("grpc conn shutdown")

	defaultClientPoolConnsSizeCap = 5
	defaultDialTimeout            = 5 * time.Second
	defaultKeepAlive              = 30 * time.Second
	defaultKeepAliveTimeout       = 10 * time.Second
)

type ClientOption struct {
	ClientPoolConnSize int
	IsTls              bool
	DialTimeOut        time.Duration
	KeepAlive          time.Duration
	KeepAliveTimeout   time.Duration
}

var (
	_ IClientPool = (*ClientPool)(nil)
)

type IClientPool interface {
	Get() (*grpc.ClientConn, error)
	ConnCount() int64
	Close()
}

type ClientPool struct {
	target   string
	option   *ClientOption
	next     int64
	cap      int64
	connCont int64
	sync.Mutex
	conns []*grpc.ClientConn
}

func (cc *ClientPool) ConnCount() int64 {
	return atomic.LoadInt64(&cc.connCont)
}

func (cc *ClientPool) Get() (*grpc.ClientConn, error) {
	return cc.getConn()
}

func (cc *ClientPool) getConn() (*grpc.ClientConn, error) {
	var (
		idx  int64
		next int64
		err  error
	)

	next = atomic.AddInt64(&cc.next, 1)
	idx = next % cc.cap
	atomic.SwapInt64(&cc.next, idx)
	conn := cc.conns[idx]
	if conn != nil && cc.checkState(conn) == nil {
		return conn, nil
	}

	//gc old conn
	if conn != nil {
		conn.Close()
		atomic.AddInt64(&cc.connCont, -1)
	}

	cc.Lock()
	defer cc.Unlock()

	//double check, Prevent have been initialized
	if conn != nil && cc.checkState(conn) == nil {
		return conn, nil
	}

	conn, err = cc.connect()
	if err != nil {
		return nil, err
	}
	cc.conns[idx] = conn
	atomic.AddInt64(&cc.connCont, 1)
	return conn, nil
}

func (cc *ClientPool) checkState(conn *grpc.ClientConn) error {
	state := conn.GetState()
	switch state {
	case connectivity.Idle, connectivity.TransientFailure, connectivity.Shutdown:
		return ErrConnShutdown
	}

	return nil
}

func (cc *ClientPool) connect() (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), cc.option.DialTimeOut)
	defer cancel()
	opts := make([]grpc.DialOption, 0, 3)
	if cc.option.IsTls {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	opts = append(opts,
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    cc.option.KeepAlive,
			Timeout: cc.option.KeepAliveTimeout,
		}),
	)

	conn, err := grpc.DialContext(ctx, cc.target, opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (cc *ClientPool) Close() {
	cc.Lock()
	defer cc.Unlock()

	for _, conn := range cc.conns {
		if conn == nil {
			continue
		}

		conn.Close()
	}
}

func NewClientPoolWithOption(target string, option *ClientOption) *ClientPool {
	if (option.ClientPoolConnSize) <= 0 {
		option.ClientPoolConnSize = defaultClientPoolConnsSizeCap
	}

	if option.DialTimeOut <= 0 {
		option.DialTimeOut = defaultDialTimeout
	}

	if option.KeepAlive <= 0 {
		option.KeepAlive = defaultKeepAlive
	}

	if option.KeepAliveTimeout <= 0 {
		option.KeepAliveTimeout = defaultKeepAliveTimeout
	}

	return &ClientPool{
		target: target,
		option: option,
		cap:    int64(option.ClientPoolConnSize),
		conns:  make([]*grpc.ClientConn, option.ClientPoolConnSize),
	}
}
