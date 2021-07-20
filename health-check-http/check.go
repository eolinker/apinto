package health_check_http

//
//type checker struct {
//	ctx        context.Context
//	cancelFunc context.CancelFunc
//	ch         chan *checkNode
//	config     *Config
//	id         string
//}
//
//func (c *checker) doCheckLoop() {
//	nodes := make(map[string]map[string]discovery.INode)
//	for {
//		select {
//		case <-c.ctx.Done():
//			{
//				return
//			}
//		case n, ok := <-c.ch:
//			{
//
//			}
//		}
//	}
//}
//
//func (c *checker) AddToCheck(node discovery.INode) error {
//	n := &checkNode{
//		node: node,
//		from: c.id,
//	}
//	c.ch <- n
//	return nil
//}
//
//func (c *checker) Stop() error {
//	c.cancelFunc()
//	return nil
//}
//
//type checkNode struct {
//	node discovery.INode
//	from string
//}
//
//func NewHealthCheck(conf interface{}) (discovery.IHealthChecker, error) {
//	ctx, cancel := context.WithCancel(context.Background())
//	ch := make(chan *checkNode, 10)
//	c := &checker{
//		ctx:        ctx,
//		cancelFunc: cancel,
//		ch:         ch,
//	}
//	return c, nil
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
