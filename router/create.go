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
	"strings"
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
	root:=newCreateRoot(helper)

	for i:=range rules{
		r:=rules[i]
		err:=root.add(r.Path,r.Target)
		if err!= nil{
			return nil,err
		}
	}
	return root.toRouter( ),nil
}

type createNodes map[string]*createNode

func (cns createNodes)add(path []RulePath,endpoint *tEndpoint)error  {
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
func (cns createNodes)toRouter(helper ICreateHelper) Routers {

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

type PathSort struct {
	path []RulePath
	helper ICreateHelper
}

func (p *PathSort) Len() int {
	return len(p.path)
}

func (p *PathSort) Less(i, j int) bool {
	return p.helper.Less(p.path[i].CMD,p.path[j].CMD)
}

func (p *PathSort) Swap(i, j int) {
	p.path[i],p.path[j]= p.path[j],p.path[i]
}

type createRoot struct {
	helper ICreateHelper
	nexts createNodes
 }

func newCreateRoot(helper ICreateHelper) *createRoot {
	return &createRoot{
		nexts: make(createNodes),
		helper:helper,
	}
}

func (cr *createRoot) toRouter() IRouter {

	return cr.nexts.toRouter(cr.helper)

}
type IEndPoint interface {
	CMDs()[]string
	Get(CMD string)(checker.Checker,bool)
	Target()string
	EndPoint()string
}
type tEndpoint struct {
	target string
	cmds []string
	checkers map[string]checker.Checker
	endpoint string
}

func (e *tEndpoint) CMDs() []string {
	return e.cmds
}

func (e *tEndpoint) Get(CMD string) (checker.Checker,bool) {
	c,h:=e.checkers[CMD]
	return c,h
}

func (e *tEndpoint) Target() string {
	return e.target
}
func (e *tEndpoint)EndPoint()string {

	return e.target
}

func NewEndpoint(target string, path []RulePath) *tEndpoint {
	cs:=make(map[string]checker.Checker)
	cmds:=make([]string,0,len(path))
	build:=strings.Builder{}

	for _,p:=range path{
		cs[p.CMD] = p.Checker
		cmds = append(cmds, p.CMD)

		build.WriteString(p.CMD)
		build.WriteString(p.Checker.Key())
		build.WriteString("&")
	}

	return &tEndpoint{target: target, checkers:cs,cmds:cmds,endpoint:build.String()}
}

func (e *tEndpoint) Router(source ISource) (endpoint IEndPoint, has bool) {
	return e,e!= nil
}

func (cr *createRoot) add(path []RulePath,target string) error {

	if len(path) == 0 || target == "" {
		return fmt.Errorf("invalid router")
	}

	cl:= &PathSort{
		path:path,
		helper:cr.helper,
	}
	sort.Sort(cl)
	return cr.nexts.add(path,	NewEndpoint(target,path))

}
type createChecker struct {
	checker checker.Checker
	nexts createNodes
	endpoint *tEndpoint
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

	 routers:= cc.nexts.toRouter(helper)
	 // if there is endpoint, append to end
	if cc.endpoint != nil{
		routers = append(routers, cc.endpoint)
	}
	return routers
}
func (cc *createChecker) add(path []RulePath,endpoint *tEndpoint) error {
 	if len(path) == 0 {
		if cc.endpoint != nil{
			return fmt.Errorf("%s: exist",endpoint.endpoint)
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
func (cn *createNode)add(checker checker.Checker,path []RulePath,endpoint *tEndpoint)  error{

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

	return len(ci.endpoint.cmds) > len(cj.endpoint.cmds)

}

func (cks createCheckers) Swap(i, j int) {
	cks[i],cks[j]= cks[j],cks[i]
}


