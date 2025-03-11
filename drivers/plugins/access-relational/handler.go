package access_relational

import (
	"strconv"
	"time"

	"github.com/eolinker/eosc"
	"github.com/eolinker/eosc/metrics"
	"github.com/eolinker/eosc/utils"
)

var (
	_ ruleHandler = (*handler)(nil)
)

type ruleHandler interface {
	Check(ctx eosc.IEntry) bool
}
type handler struct {
	a metrics.Metrics
	b metrics.Metrics
}

func (h *handler) Check(entry eosc.IEntry) bool {
	af := h.ReadA(entry)
	bf := h.ReadB(entry)

	intersection := utils.Intersection(af, bf)

	return len(intersection) > 0
}

func (h *handler) ReadA(entry eosc.IEntry) []string {
	key := h.a.Metrics(entry)
	return read(key)
}
func (h *handler) ReadB(entry eosc.IEntry) []string {
	key := h.b.Metrics(entry)
	return read(key)
}
func read(key string) []string {
	all, has := customerVar.GetAll(key)
	if !has {
		return nil
	}
	now := time.Now().UnixMilli()
	result := make([]string, 0, len(all))
	for k, v := range all {
		timestamp, _ := strconv.ParseInt(v, 10, 64)
		if timestamp <= 0 || timestamp > now {
			result = append(result, k)
		}
	}
	return result
}
