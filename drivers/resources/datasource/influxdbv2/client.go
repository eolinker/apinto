package influxdbv2

import (
	"reflect"

	monitor_entry "github.com/eolinker/apinto/entries/monitor-entry"

	"github.com/eolinker/eosc/log"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type Client struct {
	id string
	influxdb2.Client
	api.WriteAPI
}

func NewClient(cfg *Config) *Client {
	id := ""
	client := influxdb2.NewClient(cfg.Url, cfg.Token)
	writeAPI := client.WriteAPI(cfg.Org, cfg.Bucket)
	return &Client{
		id,
		client,
		writeAPI,
	}
}

func (c *Client) ID() string {
	return c.id
}

func (c *Client) Write(point monitor_entry.IPoint) error {
	if c.WriteAPI != nil {
		p, ok := point.(monitor_entry.IPoint)
		if !ok {
			log.Error("need: ", reflect.TypeOf((monitor_entry.IPoint)(nil)), "now: ", reflect.TypeOf(point))
			return nil
		}
		log.Debug("table: ", p.Table(), " tags: ", p.Tags(), " fields: ", p.Fields(), " time: ", p.Time())
		c.WritePoint(influxdb2.NewPoint(
			p.Table(),
			p.Tags(),
			p.Fields(),
			p.Time(),
		))
		return nil
	}
	return nil
}

func (c *Client) Close() {
	if c.Client != nil {
		c.Client.Close()
	}
	c.Client = nil
}
