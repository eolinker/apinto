package influxdb_v2

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eolinker/apinto/checker"
	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/apinto/output"

	"github.com/eolinker/apinto/drivers"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/eosc"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

const (
	maxRetries    = 3
	retryInterval = 100 * time.Millisecond
)

type fieldType string

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
	lock        sync.RWMutex
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
	client := NewClient(conf)
	if _, err := client.Ping(o.ctx); err != nil {
		client.Close()
		return fmt.Errorf("connect influxdbv2 error: %w", err)
	}

	o.lock.Lock()
	oldClient := o.client
	o.client = client
	o.metrics = conf.Metrics
	o.measurement = conf.Measurement
	o.conf = conf
	o.filters = parseFilters(conf.Filters)
	o.lock.Unlock()

	if oldClient != nil {
		oldClient.Close()
	}

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
	o.lock.Lock()
	client := o.client
	o.client = nil
	o.lock.Unlock()

	if client != nil {
		client.Close()
	}
	scope_manager.Del(o.Id())
	return nil
}

func (o *Output) Output(entry eosc.IEntry) error {
	o.lock.RLock()
	filters := o.filters
	measurementPattern := o.measurement
	metrics := o.metrics
	fieldsConf := o.conf.Fields
	o.lock.RUnlock()

	for _, f := range filters {
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

	//msec := eosc.ReadStringFromEntry(entry, "now")
	//msecInt, _ := strconv.ParseInt(msec, 10, 64)
	//var timestamp time.Time
	//if msecInt > 0 {
	//	timestamp = time.UnixMilli(msecInt)
	//} else {
	//
	//}
	timestamp := time.Now()

	measurement := measurementPattern
	if strings.HasPrefix(measurement, "$") {
		measurement = eosc.ReadStringFromEntry(entry, measurement[1:])
	}

	tags := make(map[string]string, len(metrics))
	for k, v := range metrics {
		if strings.HasPrefix(v, "$") {
			tags[k] = eosc.ReadStringFromEntry(entry, v[1:])
		} else {
			tags[k] = v
		}
	}

	fields := make(map[string]interface{}, len(fieldsConf))
	for k, v := range fieldsConf {
		if strings.HasPrefix(v, "$") {
			fields[k] = entry.Read(v[1:])
		} else if strings.HasPrefix(v, "#") {
			val := entry.Read(v[1:])
			switch vv := val.(type) {
			case string:
				target, _ := strconv.ParseFloat(vv, 64)
				fields[k] = target
			case int:
				fields[k] = float64(vv)
			case int64:
				fields[k] = float64(vv)
			case float64:
				fields[k] = vv
			case float32:
				fields[k] = float64(vv)
			default:
				fields[k] = val
			}

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
		return nil
	case <-o.ctx.Done():
		return o.ctx.Err()
	}
}

func (o *Output) writePointWithRetry(p *Point) {
	o.lock.RLock()
	client := o.client
	o.lock.RUnlock()

	if client == nil || client.WriteAPIBlocking == nil {
		return
	}

	point := influxdb2.NewPoint(
		p.Measurement,
		p.Tags,
		p.Fields,
		p.Time,
	)

	var err error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = client.WritePoint(o.ctx, point)
		if err == nil {
			if attempt > 1 {
				log.Infof("influxdbv2 write point to measurement %s succeeded on retry attempt %d", p.Measurement, attempt)
			} else {
				log.Debug("influxdbv2 write point succeeded, table: ", p.Measurement, " tags: ", p.Tags, " fields: ", p.Fields, " time: ", p.Time)
			}
			return
		}

		log.Warnf("influxdbv2 write point to measurement %s failed (attempt %d/%d), err: %v", p.Measurement, attempt, maxRetries, err)

		if attempt < maxRetries {
			select {
			case <-o.ctx.Done():
				return
			case <-time.After(retryInterval):
			}
		}
	}

	log.Errorf("influxdbv2 write point to measurement %s failed after %d attempts, final err: %v", p.Measurement, maxRetries, err)
}

func (o *Output) doLoop() {
	for {
		select {
		case p, ok := <-o.outputChan:
			if !ok {
				return
			}
			o.writePointWithRetry(p)
		case <-o.ctx.Done():
			return
		}
	}
}
