package strategy

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/eolinker/apinto/checker"
	"github.com/eolinker/eosc/log"
)

type IPChecker struct {
	checker IChecker
}

func (i *IPChecker) Check(v string, has bool) bool {
	if !has {
		return false
	}
	return i.checker.Check(v)
}

func (i *IPChecker) Key() string {
	return i.checker.Key()
}

func (i *IPChecker) CheckType() checker.CheckType {
	return checker.CheckTypeIP
}

func (i *IPChecker) Value() string {
	return i.checker.Value()
}

func newIPChecker(pattern string) (*IPChecker, error) {
	if strings.Contains(pattern, "*") {
		checker, err := newIpV4RangeChecker(pattern)
		if err != nil {
			return nil, err
		}
		return &IPChecker{checker}, nil
	}
	if strings.Contains(pattern, "/") {
		checker, err := newIpCidrChecker(pattern)
		if err != nil {
			return nil, err
		}
		return &IPChecker{checker}, nil
	}
	return &IPChecker{newIpEqualChecker(pattern)}, nil
}

func newIpEqualChecker(ip string) *ipEqualChecker {
	return &ipEqualChecker{ip: ip}
}

type ipEqualChecker struct {
	ip string
}

func (i *ipEqualChecker) Key() string {
	return i.ip
}

func (i *ipEqualChecker) Value() string {
	return i.ip
}

func (i *ipEqualChecker) Check(v string) bool {
	return i.ip == v
}

type ipCidrChecker struct {
	cidr *net.IPNet
	org  string
}

func (i *ipCidrChecker) Key() string {
	return i.org
}

func (i *ipCidrChecker) Value() string {
	return i.org
}

func newIpCidrChecker(ip string) (*ipCidrChecker, error) {
	_, cidr, err := net.ParseCIDR(ip)
	if err != nil {
		return nil, err
	}
	return &ipCidrChecker{cidr: cidr}, nil
}

func (i *ipCidrChecker) Check(ip string) bool {
	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		log.Error("invalid ip: %s", ip)
		return false
	}
	return i.cidr.Contains(ipAddr)
}

func newIpV4RangeChecker(ip string) (*ipV4RangeChecker, error) {
	ipParts := strings.Split(ip, ".")
	patterns := make([]string, 0, 4)
	index := 0
	for _, p := range ipParts {
		index++
		v, err := strconv.Atoi(p)
		if err != nil || v < 1 || v > 255 {
			return nil, fmt.Errorf("invalid ip: %s", ip)
		}
		patterns = append(patterns, p)
	}
	for i := 4 - index; i > 0; i-- {
		patterns = append(patterns, "*")
	}
	return &ipV4RangeChecker{ipParts: patterns}, nil
}

type ipV4RangeChecker struct {
	ipParts []string
	org     string
}

func (i *ipV4RangeChecker) Key() string {
	return i.org
}

func (i *ipV4RangeChecker) Value() string {
	return i.org
}

func (i *ipV4RangeChecker) Check(ip string) bool {
	ipParts := strings.Split(ip, ".")
	if len(ipParts) != 4 {
		return false
	}
	for index, p := range i.ipParts {
		if p == "*" {
			continue
		}
		if p != ipParts[index] {
			return false
		}
	}
	return true
}

type IChecker interface {
	Check(v string) bool
	Key() string
	Value() string
}
