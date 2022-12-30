package influxdbv2

import (
	"context"
	"reflect"

	monitor_entry "github.com/eolinker/apinto/monitor-entry"

	"github.com/eolinker/eosc/log"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type Client struct {
	id string
	influxdb2.Client
	api.WriteAPIBlocking
}

func NewClient(cfg *Config) *Client {
	id := ""
	client := influxdb2.NewClient(cfg.Url, cfg.Token)

	return &Client{
		id,
		client,
		client.WriteAPIBlocking(cfg.Org, cfg.Bucket),
	}
}

func (c *Client) ID() string {
	return c.id
}

func (c *Client) Write(point monitor_entry.IPoint) error {
	if c.WriteAPIBlocking != nil {
		p, ok := point.(monitor_entry.IPoint)
		if !ok {
			log.Error("need: ", reflect.TypeOf((monitor_entry.IPoint)(nil)), "now: ", reflect.TypeOf(point))
			return nil
		}
		return c.WritePoint(context.Background(), influxdb2.NewPoint(
			p.Table(),
			p.Tags(),
			p.Fields(),
			p.Time(),
		))
	}
	return nil
}

func (c *Client) Close() {
	if c.Client != nil {
		c.Client.Close()
	}
	c.Client = nil
}
