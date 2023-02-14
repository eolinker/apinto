package grpc_context

import (
	"sync"
	"time"
)

var _ IClient = (*Client)(nil)

var (
	clientPool          IClient = NewClient()
	defaultClientOption         = ClientOption{
		ClientPoolConnSize: defaultClientPoolConnsSizeCap,
		DialTimeOut:        defaultDialTimeout,
		KeepAlive:          defaultKeepAlive,
		KeepAliveTimeout:   defaultKeepAliveTimeout,
	}
)

type IClient interface {
	Get(target string, isTls bool) (IClientPool, bool)
	Set(target string, isTls bool, pool IClientPool)
	Del(target string, isTls bool)
	Close()
}

type Client struct {
	clients    map[string]IClientPool
	tlsClients map[string]IClientPool
	locker     sync.RWMutex
}

func NewClient() *Client {
	client := &Client{
		clients:    make(map[string]IClientPool),
		tlsClients: make(map[string]IClientPool),
		locker:     sync.RWMutex{},
	}
	go client.clean()
	return client
}

func (c *Client) clean() {
	sleep := time.Second * 10
	for {
		c.locker.Lock()
		for key, client := range c.clients {
			if client.ConnCount() < 1 {
				delete(c.clients, key)
			}
		}
		for key, client := range c.tlsClients {
			if client.ConnCount() < 1 {
				delete(c.clients, key)
			}
		}
		c.locker.Unlock()
		time.Sleep(sleep)
	}
}

func (c *Client) Get(target string, isTls bool) (IClientPool, bool) {
	c.locker.RLock()
	defer c.locker.RUnlock()
	clients := c.clients
	if isTls {
		clients = c.tlsClients
	}
	client, ok := clients[target]

	return client, ok
}

func (c *Client) Set(target string, isTls bool, pool IClientPool) {
	c.locker.Lock()
	defer c.locker.Unlock()
	c.del(target, isTls)
	clients := c.clients
	if isTls {
		clients = c.tlsClients
	}
	clients[target] = pool
}

func (c *Client) Del(target string, isTls bool) {
	c.locker.Lock()
	defer c.locker.Unlock()
	c.del(target, isTls)
}

func (c *Client) del(target string, isTls bool) {
	clients := c.clients
	if isTls {
		clients = c.tlsClients
	}
	client, ok := clients[target]
	if ok {
		client.Close()
		delete(clients, target)
	}
}

func (c *Client) Close() {
	c.locker.Lock()
	defer c.locker.Unlock()
	clients := c.clients
	tlsClients := c.tlsClients
	for _, client := range clients {
		client.Close()
	}
	for _, client := range tlsClients {
		client.Close()
	}
	c.clients = nil
	c.tlsClients = nil
}
