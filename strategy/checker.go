package strategy

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

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
	return i.cidr.String()
}

func newIpCidrChecker(ip string) (*ipCidrChecker, error) {
	_, cidr, err := net.ParseCIDR(ip)
	if err != nil {
		return nil, err
	}
	return &ipCidrChecker{cidr: cidr, org: ip}, nil
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

type timestampChecker struct {
	org       string
	startTime time.Time
	endTime   time.Time
}

func newTimestampChecker(timeRange string) (*timestampChecker, error) {
	// 正则表达式：匹配 HH:mm:ss - HH:mm:ss
	regex := `^((?:[01]\d|2[0-3]):[0-5]\d:[0-5]\d|24:00:00) - ((?:[01]\d|2[0-3]):[0-5]\d:[0-5]\d|24:00:00)$`
	re := regexp.MustCompile(regex)
	if !re.MatchString(timeRange) {
		return nil, fmt.Errorf("invalid time format, expected HH:mm:ss - HH:mm:ss (00:00:00 - 24:00:00)")
	}

	// 提取开始时间和结束时间
	times := strings.Split(timeRange, " - ")
	startTimeStr, endTimeStr := times[0], times[1]
	// 解析开始时间和结束时间（假设在当前日期）
	startTime, err := time.Parse("15:04:05", startTimeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start time: %v", err)
	}
	if endTimeStr == "24:00:00" {
		// 特殊处理 24:00:00，表示第二天的 00:00:00
		endTimeStr = "23:59:59"
	}

	endTime, err := time.Parse("15:04:05", endTimeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse end time: %v", err)
	}
	if startTime.After(endTime) {
		return nil, fmt.Errorf("start time %s cannot be after end time %s", startTimeStr, endTimeStr)
	}
	return &timestampChecker{startTime: startTime, endTime: endTime}, nil

}

func (t *timestampChecker) Check(v string, has bool) bool {
	if !has {
		return false
	}
	now, err := time.ParseInLocation("2006-01-02 15:04:05", v, time.Local)
	if err != nil {
		log.Error("invalid timestamp format: %s, error: %v", v, err)
		return false
	}
	startTime := time.Date(now.Year(), now.Month(), now.Day(), t.startTime.Hour(), t.startTime.Minute(), t.startTime.Second(), 0, time.Local)
	endTime := time.Date(now.Year(), now.Month(), now.Day(), t.endTime.Hour(), t.endTime.Minute(), t.endTime.Second(), 0, time.Local)
	if startTime.Before(now) && endTime.After(now) {
		return true
	}
	return false
}

func (t *timestampChecker) Key() string {
	return t.org
}

func (t *timestampChecker) CheckType() checker.CheckType {
	return checker.CheckTypeTimeRange
}

func (t *timestampChecker) Value() string {
	return t.org
}

type IChecker interface {
	Check(v string) bool
	Key() string
	Value() string
}
