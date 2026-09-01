package influxdb_v2

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type Client struct {
	influxdb2.Client
	api.WriteAPIBlocking
}

func NewClient(cfg *Config) *Client {
	client := influxdb2.NewClient(cfg.Url, cfg.Token)
	return &Client{
		Client:           client,
		WriteAPIBlocking: client.WriteAPIBlocking(cfg.Org, cfg.Bucket),
	}
}

func (c *Client) Close() {
	if c.Client != nil {
		c.Client.Close()
	}
	c.Client = nil
	c.WriteAPIBlocking = nil
}
