package kubernetes

import (
	"context"
	"fmt"
	"github.com/eolinker/apinto/discovery"
	"github.com/eolinker/eosc/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/url"
)

type client struct {
	client    *kubernetes.Clientset
	ctx       context.Context
	inner     bool
	name      string
	portName  string
	namespace string
}

func newClient(ctx context.Context, name string, cfg *AccessConfig) (*client, error) {
	var restCfg *rest.Config
	if cfg.Inner {
		var err error
		restCfg, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		if len(cfg.Address) < 1 {
			return nil, fmt.Errorf("address is nil")
		}
		u, err := url.Parse(cfg.Address[0])
		if err != nil {
			return nil, err
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			u.Scheme = "https"
		}
		restCfg = &rest.Config{
			Host:        u.String(),
			Username:    cfg.Username,
			Password:    cfg.Password,
			BearerToken: cfg.BearerToken,
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: true,
			},
		}
	}
	c, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, err
	}

	return &client{
		client:    c,
		name:      name,
		ctx:       ctx,
		namespace: cfg.Namespace,
		inner:     cfg.Inner,
		portName:  cfg.PortName,
	}, nil
}

// GetNodeList 从Client获取对应服务的节点列表
func (c *client) GetNodeList(serviceName string) ([]discovery.NodeInfo, error) {
	if c.inner {
		return c.getInternalAccess(serviceName)
	}
	return c.getExternalAccess(serviceName)
}

func (c *client) getInternalAccess(serviceName string) ([]discovery.NodeInfo, error) {
	endpoints, err := c.client.CoreV1().Endpoints(c.namespace).Get(c.ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get endpoints fail. err: %w", err)
	}
	if len(endpoints.Subsets) == 0 {
		return nil, fmt.Errorf("no available subset")
	}
	nodes := make([]discovery.NodeInfo, 0, 10)
	for _, subset := range endpoints.Subsets {
		for _, addr := range subset.Addresses {
			port := -1
			for _, p := range subset.Ports {
				// 筛选逻辑：匹配名称或端口号
				if p.Name == c.portName {
					port = int(p.Port)
					break
				}
			}
			// 如果该 Pod 无匹配端口，提示
			if len(subset.Ports) > 0 && port < 0 {
				port = int(subset.Ports[0].Port)
			}

			if port < 0 {
				log.Errorf("no available port, service: %s, subset: %s", serviceName, subset.String())
				continue
			}
			log.DebugF("service: %s, subset: %s, port: %d", serviceName, subset.String(), port)
			nodes = append(nodes, discovery.NodeInfo{
				Ip:   addr.IP,
				Port: port,
			})

		}
	}
	return nodes, nil
}

func (c *client) getExternalAccess(serviceName string) ([]discovery.NodeInfo, error) {
	// 获取 Nodes（用于 NodePort）
	availableNodes, err := c.client.CoreV1().Nodes().List(c.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("get nodes fail. err: %w", err)
	}
	service, err := c.client.CoreV1().Services(c.namespace).Get(c.ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get service fail.service: %s, err: %w", serviceName, err)
	}
	switch service.Spec.Type {
	case v1.ServiceTypeLoadBalancer:
		if len(service.Status.LoadBalancer.Ingress) <= 0 {
			return nil, fmt.Errorf("no available ingress")
		}
		nodes := make([]discovery.NodeInfo, 0, 10)
		for _, ingress := range service.Status.LoadBalancer.Ingress {
			ip := ingress.IP
			if ip == "" {
				ip = ingress.Hostname
			}
			port := 0
			for _, p := range service.Spec.Ports {
				// 筛选逻辑：匹配名称或端口号
				if p.Name == c.portName {
					port = int(p.Port)
					break
				}
			}
			if port == 0 {
				log.Errorf("no available port for service: %s", serviceName)
				continue
			}
			log.DebugF("service: %s, ip: %s, port: %d", serviceName, ip, port)
			nodes = append(nodes, discovery.NodeInfo{
				Ip:   ip,
				Port: port,
			})
		}

		return nodes, nil
	case v1.ServiceTypeNodePort:
		port := 0
		for _, p := range service.Spec.Ports {
			if port == 0 {
				if p.NodePort != 0 {
					port = int(p.NodePort)
					continue
				}
			}
			// 筛选逻辑：匹配名称或端口号
			if p.Name == c.portName {
				if p.NodePort != 0 {
					port = int(p.NodePort)
				}
			}
		}
		if port == 0 {
			return nil, fmt.Errorf("no available port for service: %s", serviceName)
		}
		nodes := make([]discovery.NodeInfo, 0, 10)
		for _, node := range availableNodes.Items {
			for _, addr := range node.Status.Addresses {
				switch addr.Type {
				case v1.NodeInternalIP:
					log.DebugF("node: %s, ip: %s, port: %d", node.Name, addr.Address, port)
					nodes = append(nodes, discovery.NodeInfo{
						Ip:   addr.Address,
						Port: port,
					})
				case v1.NodeExternalIP:
					log.DebugF("node: %s, ip: %s, port: %d", node.Name, addr.Address, port)
					nodes = append(nodes, discovery.NodeInfo{
						Ip:   addr.Address,
						Port: port,
					})
				}
			}
		}
		return nodes, nil

	case v1.ServiceTypeClusterIP:
		return nil, fmt.Errorf("the %s type is for internal access only. To enable external access, change the type to NodePort/LoadBalancer or use Ingress", service.Spec.Type)
	default:
		return nil, fmt.Errorf("unsupported service type: %s", service.Spec.Type)
	}
}

func (c *client) Close() error {
	return nil
}
