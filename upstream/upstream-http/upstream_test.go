package upstream_http

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/eolinker/eosc"

	"github.com/eolinker/goku-eosc/discovery"

	"github.com/go-basic/uuid"

	http_context "github.com/eolinker/goku-eosc/node/http-context"
	"github.com/eolinker/goku-eosc/upstream"
)

var (
	errorAppConfig = errors.New("error app config")
	s              = &Service{
		name:    "参数打印",
		desc:    "打印所有参数",
		retry:   3,
		timeout: time.Second * 10,
		scheme:  "http",
	}
	nodeConfigs = []node{
		{
			label: map[string]string{
				"weight": "10",
				"desc":   "本地",
			},
			ip:   "127.0.0.1",
			port: 8580,
		},
		{
			label: map[string]string{
				"weight": "10",
				"desc":   "链接失败",
			},
			ip:   "10.1.0.1",
			port: 8080,
		},
	}
	configs = []*TestConfig{
		{
			cfg: &Config{
				Desc:      "测试协议不正确",
				Scheme:    "tcp",
				Type:      "round-robin",
				Config:    "static_1",
				Discovery: "static_1",
			},
			name: "测试协议不正确",
			err:  errorScheme,
		},
		{
			cfg:  nil,
			err:  ErrorStructType,
			name: "配置为空",
		},
		{
			cfg: &Config{
				Desc:      "测试负载算法不存在",
				Scheme:    "http",
				Type:      "round-robin1",
				Config:    "static_1",
				Discovery: "static_1",
			},
			name: "测试负载算法不存在，负载算法不存在时不会报错",
			err:  nil,
		},
		{
			cfg: &Config{
				Desc:      "数据正常",
				Scheme:    "http",
				Type:      "round-robin",
				Config:    "static_1",
				Discovery: "static_1",
			},
			name: "数据正常",
			err:  nil,
		},
		{
			cfg: &Config{
				Desc:      "discovery不存在",
				Scheme:    "http",
				Type:      "round-robin",
				Config:    "static_1",
				Discovery: "static",
			},
			name: "discovery不存在",
			err:  errorCreateWorker,
		},
		{
			cfg: &Config{
				Desc:      "配置解析失败",
				Scheme:    "http",
				Type:      "round-robin",
				Config:    "static",
				Discovery: "static_1",
			},
			name: "配置解析失败",
			err:  errorAppConfig,
		},
	}
)

type TestConfig struct {
	cfg  *Config
	name string
	err  error
}

type Service struct {
	name    string
	desc    string
	retry   int
	timeout time.Duration
	scheme  string
	addr    string
}

func (s *Service) Name() string {
	return s.name
}

func (s *Service) Desc() string {
	return s.desc
}

func (s *Service) Retry() int {
	return s.retry
}

func (s *Service) Timeout() time.Duration {
	return s.timeout
}

func (s *Service) Scheme() string {
	return s.scheme
}

func (s *Service) ProxyAddr() string {
	return s.ProxyAddr()
}

type discoveryDemo struct {
}

func (d *discoveryDemo) Id() string {
	panic("implement me")
}

func (d *discoveryDemo) Start() error {
	panic("implement me")
}

func (d *discoveryDemo) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	panic("implement me")
}

func (d *discoveryDemo) Stop() error {
	panic("implement me")
}

func (d *discoveryDemo) CheckSkill(skill string) bool {
	panic("implement me")
}

func (d *discoveryDemo) Remove(id string) error {
	return nil
}

func (d *discoveryDemo) GetApp(config string) (discovery.IApp, error) {
	if config != "static_1" {
		return nil, errorAppConfig
	}
	nodes := map[string]discovery.INode{}
	for _, n := range nodeConfigs {
		id := uuid.New()
		nodes[id] = discovery.NewNode(n.label, id, n.ip, n.port)
	}
	attrs := map[string]string{
		"scheme": "http",
	}
	app := discovery.NewApp(nil, d, attrs, nodes)
	return app, nil
}

type node struct {
	label map[string]string
	ip    string
	port  int
}

func TestUpstream(t *testing.T) {
	f := NewFactory()
	driver, err := f.Create("upstream", "http", "", "", nil)
	if err != nil {
		t.Error(err)
		return
	}
	dDemo := &discoveryDemo{}
	allWorker := map[eosc.RequireId]interface{}{
		"static_1": dDemo,
	}

	r, err := http.NewRequest("POST", "http://localhost:8080/Web/Test/body/print", strings.NewReader("eolinker"))
	if err != nil {
		t.Error(ErrorStructType)
		return
	}
	ctx := http_context.NewContext(r, &response{})
	// 设置目标URL
	ctx.ProxyRequest.SetTargetURL(r.URL.Path)

	for _, c := range configs {
		t.Run(c.name, func(t *testing.T) {
			worker, err := driver.Create(uuid.New(), "upstream", c.cfg, allWorker)
			if err != nil {
				if c.err != nil && strings.Contains(err.Error(), c.err.Error()) {
					return
				}

				t.Error(err)
				return
			}
			u, _ := worker.(upstream.IUpstream)
			for i := 0; i < 10; i++ {
				resp, err := u.Send(ctx, s)
				if err != nil {
					t.Error(err)
					continue
				}
				if string(resp.Body()) != "eolinker" {
					t.Error("error response")
					continue
				}
			}

		})
	}

}

type response struct {
}

func (r *response) Header() http.Header {
	panic("implement me")
}

func (r *response) Write(bytes []byte) (int, error) {
	panic("implement me")
}

func (r *response) WriteHeader(statusCode int) {
	panic("implement me")
}
