package influxdb_v2

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/apinto/output"

	"github.com/eolinker/apinto/drivers"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/eosc"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

var _ output.IEntryOutput = (*Output)(nil)
var _ eosc.IWorker = (*Output)(nil)

type Point struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]interface{}
	Time        time.Time
}

type Output struct {
	drivers.WorkerBase
	client      *Client
	metrics     map[string]string
	measurement string
	outputChan  chan *Point
	ctx         context.Context
	cancel      context.CancelFunc
	conf        *Config
}

func (o *Output) Start() error {
	return nil
}

func (o *Output) Reset(conf interface{}, workers map[eosc.RequireId]eosc.IWorker) error {
	cfg, err := check(conf)
	if err != nil {
		return err
	}
	if err := o.reset(cfg); err != nil {
		return err
	}
	return nil
}

func (o *Output) reset(conf *Config) error {
	if reflect.DeepEqual(conf, o.conf) {
		return nil
	}

	client := NewClient(conf)
	if _, err := client.Ping(o.ctx); err != nil {
		client.Close()
		return fmt.Errorf("connect influxdbv2 error: %w", err)
	}

	if o.client != nil {
		o.client.Close()
	}

	o.client = client
	o.metrics = conf.Metrics
	o.measurement = conf.Measurement
	o.conf = conf

	scope_manager.Set(o.Id(), o, conf.Scopes...)
	return nil
}

func (o *Output) Stop() error {
	o.Close()

	return nil
}

func (o *Output) CheckSkill(skill string) bool {
	return output.CheckSkill(skill)
}

func (o *Output) Close() error {
	o.cancel()
	if o.client != nil {
		o.client.Close()
		o.client = nil
	}
	scope_manager.Del(o.Id())
	return nil
}

func (o *Output) Output(entry eosc.IEntry) error {
	msec := eosc.ReadStringFromEntry(entry, "msec")
	msecInt, _ := strconv.ParseInt(msec, 10, 64)
	var timestamp time.Time
	if msecInt > 0 {
		timestamp = time.UnixMilli(msecInt)
	} else {
		timestamp = time.Now()
	}

	measurement := o.measurement
	if strings.HasPrefix(measurement, "$") {
		measurement = eosc.ReadStringFromEntry(entry, measurement[1:])
	}

	tags := make(map[string]string)
	for k, v := range o.metrics {
		if strings.HasPrefix(v, "$") {
			tags[k] = eosc.ReadStringFromEntry(entry, v[1:])
		} else {
			tags[k] = v
		}
	}

	fields := make(map[string]interface{})
	for k, v := range o.conf.Fields {
		if s, ok := v.(string); ok && strings.HasPrefix(s, "$") {
			fields[k] = entry.Read(s[1:])
		} else {
			fields[k] = v
		}
	}

	if len(fields) == 0 {
		return nil
	}

	select {
	case o.outputChan <- &Point{
		Measurement: measurement,
		Tags:        tags,
		Fields:      fields,
		Time:        timestamp,
	}:
	default:
		log.Warn("influxdbv2 output channel is full, drop point")
	}

	return nil
}

func (o *Output) doLoop() {
	for {
		select {
		case p, ok := <-o.outputChan:
			if !ok {
				return
			}
			if o.client == nil || o.client.WriteAPI == nil {
				continue
			}
			o.client.WritePoint(influxdb2.NewPoint(
				p.Measurement,
				p.Tags,
				p.Fields,
				p.Time,
			))
		case <-o.ctx.Done():
			return
		}
	}
}
