package loki

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/eolinker/eosc/formatter"

	scope_manager "github.com/eolinker/apinto/scope-manager"

	"github.com/eolinker/apinto/output"

	"github.com/eolinker/apinto/drivers"

	"github.com/eolinker/eosc/log"

	"github.com/eolinker/eosc"
)

var _ output.IEntryOutput = (*Output)(nil)
var _ eosc.IWorker = (*Output)(nil)

var (
	client = http.Client{}
)

type Request struct {
	Streams []*Stream `json:"streams"`
}

type Stream struct {
	Stream map[string]string `json:"stream"`
	Values [][]interface{}   `json:"values"`
}

type Output struct {
	drivers.WorkerBase
	url        string
	method     string
	headers    map[string]string
	labels     map[string]string
	formatter  eosc.IFormatter
	outputChan chan *Request
	ctx        context.Context
	cancel     context.CancelFunc
	conf       *Config
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
	//创建formatter
	factory, has := formatter.GetFormatterFactory(conf.Type)
	if !has {
		return fmt.Errorf("formatter %s not found", conf.Type)
	}
	fm, err := factory.Create(conf.Formatter)
	if err != nil {
		return fmt.Errorf("create formatter error: %v", err)
	}

	o.url = conf.Url
	o.method = conf.Method
	o.headers = conf.Headers
	o.labels = conf.Labels
	o.conf = conf
	o.formatter = fm

	scope_manager.Set(o.Id(), o, conf.Scopes...)
	return nil
}

func (o *Output) Stop() error {
	o.Close()
	o.formatter = nil

	return nil
}

func (o *Output) CheckSkill(skill string) bool {
	return output.CheckSkill(skill)
}

func (o *Output) Close() error {
	o.cancel()
	close(o.outputChan)
	return nil
}

func (o *Output) Output(entry eosc.IEntry) error {
	if o.formatter == nil {
		return nil
	}
	data := o.formatter.Format(entry)
	msec := eosc.ReadStringFromEntry(entry, "msec")
	msecInt, _ := strconv.ParseInt(msec, 10, 64)
	labels := make(map[string]string)
	for k, v := range o.labels {
		if strings.HasPrefix(v, "$") {
			labels[k] = eosc.ReadStringFromEntry(entry, v[1:])
		} else {
			labels[k] = v
		}
	}
	o.outputChan <- &Request{
		Streams: []*Stream{
			{
				Stream: labels,
				Values: [][]interface{}{
					{strconv.FormatInt(time.UnixMilli(msecInt).UnixNano(), 10), string(data)},
				},
			},
		},
	}
	return nil
}

func (o *Output) genRequest(data []byte) (*http.Request, error) {

	req, err := http.NewRequest(o.method, o.url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	for k, v := range o.headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("content-type", "application/json")
	return req, nil
}

func (o *Output) doLoop() {
	for {
		select {
		case entry, ok := <-o.outputChan:
			if !ok {
				return
			}
			data, _ := json.Marshal(entry)
			log.Infof("send data to loki: %s", string(data))
			req, err := o.genRequest(data)
			if err != nil {
				log.Errorf("gen request error: %v,data is %s", err, string(data))
				continue
			}
			resp, err := client.Do(req)
			if err != nil {
				log.Errorf("send request error: %v,data is %s", err, string(data))
				continue
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Errorf("read body error: %v,data is %s", err, string(data))
				resp.Body.Close()
				continue
			}
			resp.Body.Close()
			if resp.StatusCode > 299 {
				log.Errorf("response status error: %s,data is %s,response body is %s", resp.Status, string(data), string(body))
				continue
			}

		case <-o.ctx.Done():
			return
		}
	}
}
