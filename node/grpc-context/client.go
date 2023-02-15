package grpc_context

import (
	"fmt"
	"sync"
	"time"
)

var _ IClient = (*Client)(nil)

var (
	clientPool IClient = NewClient()
)

type IClient interface {
	Get(target string, isTls bool, host ...string) IClientPool
	Close()
}

type Client struct {
	clients    map[string]IClientPool
	tlsClients map[string]IClientPool
	stop       bool
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
		if c.stop {
			c.locker.Unlock()
			return
		}
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

func (c *Client) Get(target string, isTls bool, host ...string) IClientPool {
	key := target
	authority := ""
	if len(host) > 0 && host[0] != "" {
		key = fmt.Sprintf("%s|%s", target, host[0])
		authority = host[0]
	}
	c.locker.RLock()
	clients := c.clients
	if isTls {
		clients = c.tlsClients
	}
	client, ok := clients[key]
	c.locker.RUnlock()
	if ok {
		return client
	}
	c.locker.Lock()
	defer c.locker.Unlock()
	client, ok = clients[key]
	if ok {
		return client
	}
	p := NewClientPoolWithOption(target, &ClientOption{
		ClientPoolConnSize: defaultClientPoolConnsSizeCap,
		DialTimeOut:        defaultDialTimeout,
		KeepAlive:          defaultKeepAlive,
		KeepAliveTimeout:   defaultKeepAliveTimeout,
		IsTls:              isTls,
		Authority:          authority,
	})

	clients[key] = p

	return p
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
	c.stop = true
	c.clients = nil
	c.tlsClients = nil
}
