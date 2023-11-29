package static

import (
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/eolinker/apinto/discovery"
	health_check_http "github.com/eolinker/apinto/health-check-http"
)

type HeathCheckHandler struct {
	healthOn bool
	checker  *health_check_http.HTTPCheck
	nodes    discovery.INodes
}

func NewHeathCheckHandler(nodes discovery.INodes, cfg *Config) *HeathCheckHandler {
	h := &HeathCheckHandler{
		nodes: nodes,
	}
	_ = h.reset(cfg)
	return h
}

func (s *HeathCheckHandler) reset(cfg *Config) error {

	s.healthOn = cfg.HealthOn
	s.nodes.SetHealthCheck(s.healthOn)
	if !cfg.HealthOn {
		checker := s.checker
		if checker != nil {
			s.checker = nil
			checker.Stop()
		}
		return nil
	}
	checker := s.checker
	if checker == nil {
		checker = health_check_http.NewHTTPCheck(
			health_check_http.Config{
				Protocol:    cfg.Health.Scheme,
				Method:      cfg.Health.Method,
				URL:         cfg.Health.URL,
				SuccessCode: cfg.Health.SuccessCode,
				Period:      time.Duration(cfg.Health.Period) * time.Second,
				Timeout:     time.Duration(cfg.Health.Timeout) * time.Millisecond,
			})
		checker.Check(s.nodes)
	} else {
		_ = checker.Reset(
			health_check_http.Config{
				Protocol:    cfg.Health.Scheme,
				Method:      cfg.Health.Method,
				URL:         cfg.Health.URL,
				SuccessCode: cfg.Health.SuccessCode,
				Period:      time.Duration(cfg.Health.Period) * time.Second,
				Timeout:     time.Duration(cfg.Health.Timeout) * time.Millisecond,
			},
		)
	}
	s.checker = checker

	return nil
}
func (s *HeathCheckHandler) stop() {

	checker := s.checker
	if checker != nil {
		s.checker = nil
		checker.Stop()
	}
}

func fields(str string) []string {
	words := strings.FieldsFunc(strings.Join(strings.Split(str, ";"), " ; "), func(r rune) bool {
		return unicode.IsSpace(r)
	})
	return words
}

// validIP 判断ip是否合法
func validIP(ip string) bool {
	match, err := regexp.MatchString(`^(?:(?:1[0-9][0-9]\.)|(?:2[0-4][0-9]\.)|(?:25[0-5]\.)|(?:[1-9][0-9]\.)|(?:[0-9]\.)){3}(?:(?:1[0-9][0-9])|(?:2[0-4][0-9])|(?:25[0-5])|(?:[1-9][0-9])|(?:[0-9]))$`, ip)
	if err != nil {
		return false
	}
	return match
}
