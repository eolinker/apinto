package health_check_http

//var supportMethods = []string{
//	"POST", "GET", "PUT", "DELETE", "OPTIONS", "HEAD", "OPTIONS",
//}
//
//type checkerFactory struct {
//	c          chan *checkNode
//	ctx        context.Context
//	cancelFunc context.CancelFunc
//}
//
//type checkNode struct {
//	checkNode discovery.INode
//	from      string
//	IsRemove  bool
//}
//
//func NewCheckerFactory() *checkerFactory {
//
//	ctx, cancelFunc := context.WithCancel(context.Background())
//	ch := make(chan *checkNode, 10)
//
//	c := checkerFactory{
//		c:          ch,
//		ctx:        ctx,
//		cancelFunc: cancelFunc,
//	}
//	go c.doCheckLoop()
//	return &c
//}
//func (c *checkerFactory) doCheckLoop() {
//	for {
//		select {
//		case <-c.ctx.Done():
//			{
//				return
//			}
//		case checkNode, ok := <-c.c:
//			{
//
//			}
//
//		}
//	}
//}
//func (c *checkerFactory) Create(config interface{}) (discovery.IHealthChecker, error) {
//	cfg, ok := config.(*Config)
//	if !ok {
//		return nil, errors.New("fail to create health checker")
//	}
//	validMethod := false
//	for _, method := range supportMethods {
//		if method == cfg.Method {
//			validMethod = true
//			break
//		}
//	}
//	if !validMethod {
//		return nil, errors.New("error request method")
//	}
//	return &checker{config: cfg, client: &http.Client{}}, nil
//}
//
//type checker struct {
//	id     string
//	config *Config
//	client *http.Client
//	nodes  map[string]*checkNode
//
//	ch chan *checkNode
//}
//
//func (c *checker) AddToCheck(checkNode discovery.INode) error {
//	n := &checkNode{
//		checkNode: checkNode,
//		from:      c.id,
//	}
//	c.ch <- n
//	c.nodes[checkNode.ID()] = n
//	return nil
//}
//
//func (c *checker) Stop() error {
//
//	return nil
//}
//
//func (c *checker) check() error {
//	for _, checkNode := range c.nodes {
//		uri := fmt.Sprintf("%s://%s/%s", c.config.Protocol, strings.TrimSuffix(checkNode.checkNode.Addr(), "/"), strings.TrimPrefix(c.config.URL, "/"))
//		c.client.Timeout = c.config.Timeout
//		request, err := http.NewRequest(c.config.Method, uri, nil)
//		if err != nil {
//			return err
//		}
//		resp, err := c.client.Do(request)
//		if err != nil {
//			return err
//		}
//		defer resp.Body.Close()
//		if c.config.SuccessCode != resp.StatusCode {
//			return errors.New("error status code")
//		}
//
//	}
//	return nil
//}
