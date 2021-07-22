package upstream_http

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/eolinker/eosc"
	http_context "github.com/eolinker/goku-eosc/node/http-context"
	"github.com/eolinker/goku-eosc/service"
	"github.com/eolinker/goku-eosc/upstream"
)

var (
	ErrorStructType = errors.New("error struct type")
)

var s = &Service{
	name:    "参数打印",
	desc:    "打印所有参数",
	retry:   3,
	timeout: time.Second * 10,
	scheme:  "http",
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

func TestUpstream(t *testing.T) {

}

func getWorker(factory eosc.IProfessionDriverFactory, cfg interface{}, profession string, name string, label string, desc string, params map[string]string, workerID, workerName string, worker map[eosc.RequireId]interface{}) (eosc.IWorker, error) {
	driver, err := factory.Create(profession, name, label, desc, params)
	if err != nil {
		return nil, err
	}

	return driver.Create(workerID, workerName, cfg, worker)
}

func send(ctx *http_context.Context, s service.IServiceDetail, hUpstream upstream.IUpstream) error {
	resp, err := hUpstream.Send(ctx, s)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(body))
	return nil
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
