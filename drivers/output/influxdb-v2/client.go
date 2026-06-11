package influxdb_v2

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type Client struct {
	influxdb2.Client
	api.WriteAPI
}

func NewClient(cfg *Config) *Client {
	client := influxdb2.NewClient(cfg.Url, cfg.Token)
	writeAPI := client.WriteAPI(cfg.Org, cfg.Bucket)
	return &Client{
		client,
		writeAPI,
	}
}

func (c *Client) Close() {
	if c.Client != nil {
		c.Client.Close()
	}
	c.Client = nil
}
