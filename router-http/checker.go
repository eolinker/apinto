package router

import "regexp"

type IChecker interface {
	Check(v string,has bool) bool
	Len()int
	Sort()int
}

func parseChecker(r string) IChecker {
	// todo parse checker
 	return nil
}
type EqualChecker map[string]bool

func (e EqualChecker) Check(v string, has bool) bool {
	if has{
		_,ok:=e[v]
		return ok
	}
	return false
}

func (e EqualChecker) Len() int {
	return 0
}

func (e EqualChecker) Sort() int {
	return -1
}

// 前缀匹配
type Prefix string

func (p Prefix) Check(v string, has bool) bool {
	panic("implement me")
}

func (p Prefix) Len() int {
	panic("implement me")
}

func (p Prefix) Sort() int {
	panic("implement me")
}

// 正则匹配
type Regexp struct {
	rule string
	*regexp.Regexp
}
type Indistinct string


type Checkers []IChecker

func (I *Checkers) Len() int {
	panic("implement me")
}

func (I *Checkers) Less(i, j int) bool {
	panic("implement me")
}

func (I *Checkers) Swap(i, j int) {
	panic("implement me")
}
