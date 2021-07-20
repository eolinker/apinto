package discover_eureka

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/eolinker/eosc"
	"github.com/eolinker/goku-eosc/discovery"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type eureka struct {
	id           string
	name         string
	address      []string
	params       map[string]string
	labels       map[string]string
	services     discovery.IServices
	context      context.Context
	cancelFunc   context.CancelFunc
}

func (e *eureka) GetApp(serviceName string) (discovery.IApp, error) {
	app, err := e.Create(serviceName)
	if err != nil {
		return nil, err
	}
	err = e.services.Set(serviceName, app.Id(), app)
	if err != nil {
		return nil, err
	}
	return app, nil
}
func (e *eureka) Create(serviceName string) (discovery.IApp, error) {
	nodes, err := e.GetNodeList(serviceName)
	if err != nil {
		return nil, err
	}
	attrs := make(discovery.Attrs)
	app := discovery.NewApp(nil, e, attrs, nodes)
	return app, nil
}

func (e *eureka) Remove(id string) error {
	return e.services.Remove(id)
}

func (e *eureka) Id() string {
	return e.id
}

func (e *eureka) Start() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	e.context = ctx
	e.cancelFunc = cancelFunc
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
	EXIT:
		for {
			select {
			case <-ctx.Done():
				break EXIT
			case <-ticker.C:
				{
					keys := e.services.AppKeys()
					for _, serviceName := range keys {
						res, err := e.GetNodeList(serviceName)
						if err != nil {
							continue
						}
						nodes := make([]discovery.INode, len(res))
						for _, v := range res {
							nodes = append(nodes, v)
						}
						e.services.Update(serviceName, nodes)
					}
				}
			}

		}
	}()
	return nil
}

func (e *eureka) Reset(conf interface{}, workers map[eosc.RequireId]interface{}) error {
	cfg, ok := conf.(*Config)
	if !ok {
		return fmt.Errorf("need %s,now %s:%w", eosc.TypeNameOf((*Config)(nil)), eosc.TypeNameOf(conf), eosc.ErrorStructType)
	}
	e.address = cfg.Config.Address
	e.params = cfg.Config.Params
	e.labels = cfg.Labels
	return nil
}

func (e *eureka) Stop() error {
	e.cancelFunc()
	return nil
}

func (e *eureka) CheckSkill(skill string) bool {
	return discovery.CheckSkill(skill)
}

func (e *eureka) GetNodeList(serviceName string) (map[string]discovery.INode, error) {
	nodes := make(map[string]discovery.INode)
	for _, addr := range e.address {
		// 获取每个ip中指定服务名的实例列表
		app, err := e.GetApplication(addr, serviceName)
		if err != nil {
			return nil, err
		}
		for _, ins := range app.Instances {
			if ins.Status != EurekaStatusUp {
				continue
			}
			port := 0
			if ins.Port.Enabled {
				port = ins.Port.Port
			} else if ins.SecurePort.Enabled {
				port = ins.SecurePort.Port
			}
			label := map[string]string{
				"app":      ins.App,
				"hostName": ins.HostName,
			}
			//for k, v := range ins.Metadata {
			//	label[k] = v
			//}
			node := discovery.NewNode(label, ins.InstanceID, ins.IPAddr, port)
			if _, ok := nodes[node.Id()]; ok {
				continue
			}
			nodes[node.Id()] = node
		}
	}
	return nodes, nil
}

func (e *eureka) GetApplication(addr, serviceName string) (*Application, error) {

	if !strings.Contains(addr, "http://") && !strings.Contains(addr, "https://") {
		addr = fmt.Sprintf("http://%s", addr)
		if v, ok := e.labels["schema"]; ok {
			if v == "https" {
				addr = fmt.Sprintf("https://%s", addr)
			}
		}
	}
	addr = fmt.Sprintf("%s/apps/%s", addr, serviceName)
	res, err := http.Get(addr)
	if err != nil {
		return nil, err
	}
	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = res.Body.Close()
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, err
	}
	var application = &Application{}
	err = xml.Unmarshal(respBody, application)
	if err != nil {
		return nil, err
	}
	return application, err
}
