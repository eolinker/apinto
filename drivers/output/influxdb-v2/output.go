package influxdb_v2

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/eolinker/apinto/checker"
	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/apinto/output"

	"github.com/eolinker/apinto/drivers"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/eosc"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

var _ output.IEntryOutput = (*Output)(nil)
var _ eosc.IWorker = (*Output)(nil)

type filter struct {
	key string
	checker.Checker
}

func parseFilters(filters []*Filter) []*filter {
	result := make([]*filter, 0, len(filters))
	for _, f := range filters {
		c, err := checker.Parse(f.Value)
		if err != nil {
			log.Errorf("parse filter value(%s) error: %v", f.Value, err)
			continue
		}
		key := f.Key
		if strings.HasPrefix(key, "$") {
			key = key[1:]
		}
		result = append(result, &filter{
			key:     key,
			Checker: c,
		})
	}
	return result
}

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
	filters     []*filter
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
	//if reflect.DeepEqual(conf, o.conf) {
	//	return nil
	//}

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
	o.filters = parseFilters(conf.Filters)

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
	for _, f := range o.filters {
		val := entry.Read(f.key)
		var checkVal string
		switch v := val.(type) {
		case string:
			checkVal = v
		case bool:
			checkVal = strconv.FormatBool(v)
		case int:
			checkVal = strconv.Itoa(v)
		case int64:
			checkVal = strconv.FormatInt(v, 10)
		case nil:
			checkVal = ""
		default:
			checkVal = fmt.Sprint(v)
		}
		if !f.Check(checkVal, true) {
			return nil
		}
	}

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
		if strings.HasPrefix(v, "$") {
			fields[k] = entry.Read(v[1:])
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
			//if c.WriteAPI != nil {
			//	p, ok := point.(monitor_entry.IPoint)
			//	if !ok {
			//		log.Error("need: ", reflect.TypeOf((monitor_entry.IPoint)(nil)), "now: ", reflect.TypeOf(point))
			//		return nil
			//	}
			log.Debug("table: ", p.Measurement, " tags: ", p.Tags, " fields: ", p.Fields, " time: ", p.Time)

			o.client.WritePoint(influxdb2.NewPoint(
				p.Measurement,
				p.Tags,
				p.Fields,
				p.Time,
			))
			o.client.WriteAPI.Flush()
		case err := <-o.client.WriteAPI.Errors():
			log.Error("influxdbv2 write error: ", err)
		case <-o.ctx.Done():
			return
		}
	}
}
