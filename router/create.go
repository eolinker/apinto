/*
 * Copyright (c) 2021. Lorem ipsum dolor sit amet, consectetur adipiscing elit.
 * Morbi non lorem porttitor neque feugiat blandit. Ut vitae ipsum eget quam lacinia accumsan.
 * Etiam sed turpis ac ipsum condimentum fringilla. Maecenas magna.
 * Proin dapibus sapien vel ante. Aliquam erat volutpat. Pellentesque sagittis ligula eget metus.
 * Vestibulum commodo. Ut rhoncus gravida arcu.
 */

package router

import (
	"fmt"
	"github.com/eolinker/goku-eosc/router/checker"
	"sort"
)

type RulePath struct {
	CMD string
 	Checker checker.Checker
}
type Rule struct {
	Path []RulePath
	Target string
}

type ICreateHelper interface {
	Less(i,j string)bool
}

//ParseRouter parse rule to IRouter
func ParseRouter(rules []Rule,helper ICreateHelper)(IRouter,error)  {
	root:=newCreateRoot()

	for i:=range rules{
		r:=rules[i]
		err:=root.add(r.Path,NewEndpoint(r.Target,r.Path))
		if err!= nil{
			return nil,err
		}
	}
	return root.toRouter(helper ),nil
}

type createNodes map[string]*createNode

func (cns createNodes)add(path []RulePath,endpoint *Endpoint)error  {
	p:=path[0]
	node,has:=cns[p.CMD]
	if !has{
		node = newCreateNode(p.CMD)
		cns[p.CMD] = node
	}
	return node.add(p.Checker,path[1:],endpoint)
}
func (cns createNodes) list(helper ICreateHelper)[]*createNode  {
	res:=make([]*createNode,0,len(cns))
	for _,v:=range cns{
		res = append(res, v)
	}
	cl:= &createNodeList{
		nodes:res,
		ICreateHelper:helper,
	}
	sort.Sort(cl)
	return cl.nodes
}
func (cns createNodes)toRouter(helper ICreateHelper) IRouter {

	nodeList := cns.list(helper)

	rl:=make([]IRouter,0,len(nodeList))

	for _,n:=range nodeList{
		r:=n.toRouter(helper)
		if r!=nil{
			rl = append(rl, r)
		}
	}
	if len(rl) == 0{
		return nil
	}
	return Routers(rl)
}
type createNodeList struct {
	nodes []*createNode

	ICreateHelper
}

func (cl *createNodeList) Len() int {
	return len(cl.nodes)
}

func (cl *createNodeList) Less(i, j int) bool {
	return cl.ICreateHelper.Less(cl.nodes[i].cmd,cl.nodes[j].cmd)
}

func (cl *createNodeList) Swap(i, j int) {
	cl.nodes[i],cl.nodes[j] = cl.nodes[j],cl.nodes[i]
}

type createRoot struct {
	nexts createNodes
 }

func newCreateRoot() *createRoot {
	return &createRoot{
		nexts: make(createNodes),
	}
}

func (cr *createRoot) toRouter(helper ICreateHelper) IRouter {

	return cr.nexts.toRouter(helper)

}
type Endpoint struct {
	target string
	path []RulePath
}

func NewEndpoint(target string, path []RulePath) *Endpoint {
	return &Endpoint{target: target, path: path}
}

func (e *Endpoint) Router(source ISource) (endpoint *Endpoint, has bool) {
	return e,e!= nil
}

func (cr *createRoot) add(path []RulePath,endpoint *Endpoint) error {
	if len(path) == 0 || endpoint == nil {
		return fmt.Errorf("invalid router")
	}
	return cr.nexts.add(path,endpoint)

}
type createChecker struct {
	checker checker.Checker
	nexts createNodes
	endpoint *Endpoint
}

func newCreateChecker(checker checker.Checker) *createChecker {
	return &createChecker{
		checker:    checker,
		nexts: 		make(createNodes),
		endpoint:     nil,
	}
}
func (cc *createChecker) toRouter(helper ICreateHelper) IRouter {

	if len(cc.nexts) == 0{
		if cc.endpoint != nil{
			return cc.endpoint
		}
		return nil
	}

	return cc.nexts.toRouter(helper)
}
func (cc *createChecker) add(path []RulePath,endpoint *Endpoint) error {
 	if len(path) == 0 {
		if cc.endpoint != nil{
			return fmt.Errorf("%s target %s: exist",cc.checker.Key(),endpoint.target)
		}
		cc.endpoint = endpoint
		return nil
	}
	return cc.nexts.add(path,endpoint)
}

type createNode struct {
	cmd string
	checkers map[string]*createChecker
}

func newCreateNode(cmd string) *createNode {
	return &createNode{
		cmd:      cmd,
		checkers: make(map[string]*createChecker),
	}
}
func (cn *createNode) toRouter(helper ICreateHelper)IRouter  {
	equals :=make(map[string]IRouter)
	tmp :=make([]*createChecker,0,len(cn.checkers))

	for _,c:=range cn.checkers{
		if c.checker.CheckType() == checker.CheckTypeEqual{
			r:= c.toRouter(helper)
			if r!= nil{
				equals[c.checker.Value()] = r
			}
		}else{
			tmp = append(tmp, c)
		}
	}
	sort.Sort(createCheckers(tmp))

	rs:=make([]IRouter,0,len(tmp))
	cs:=make([]checker.Checker,0,len(tmp))
	for _,c:=range tmp{
		r:= c.toRouter(helper)
		if r!= nil{
			rs = append(rs, r )
			cs = append(cs, c.checker)
		}
	}

	return &Node{
		cmd:cn.cmd,
		equals:equals,
		nodes:rs,
		checkers:cs,
	}
}
func (cn *createNode)add(checker checker.Checker,path []RulePath,endpoint *Endpoint)  error{

	k:=checker.Key()
	cc,has:=cn.checkers[k]
	if !has{
		cc = newCreateChecker(checker)
		cn.checkers[k] = cc
	}

	return cc.add(path,endpoint)
}



type createCheckers []*createChecker

func (cks createCheckers) Len() int {
	return len(cks)
}

func (cks createCheckers) Less(i, j int) bool {
	ci,cj := cks[i],cks[j]
	if  ci.checker.CheckType() != cj.checker.CheckType(){
		return  ci.checker.CheckType() < cj.checker.CheckType()
	}
	vl:= len(ci.checker.Value()) - len(cj.checker.Value())
	if vl != 0{
		return vl > 0
	}
	if ci.checker.Value() != cj.checker.Value(){
		return ci.checker.Value() < cj.checker.Value()
	}

	ls := len(cj.nexts)- len(cj.nexts)
	if ls != 0{
		return ls > 0
	}

	return len(ci.endpoint.path) > len(cj.endpoint.path)

}

func (cks createCheckers) Swap(i, j int) {
	cks[i],cks[j]= cks[j],cks[i]
}


