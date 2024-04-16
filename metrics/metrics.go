package metrics

import (
	"fmt"
	"github.com/eolinker/eosc/metrics"
	"strings"
)

type Metrics = metrics.Metrics

func Parse(keys []string) Metrics {

	bs := make([]string, 0, len(keys))
	for _, k := range keys {
		l := len(k)
		if l == 0 {
			continue
		}
		if len(k) >= 2 {
			if k[0] == '{' && k[l-1] == '}' {
				r := k[1 : l-1]
				if len(r) == 0 {
					continue
				}
				bs = append(bs, fmt.Sprintf("${%s}", r))
				continue
			}
		}
		bs = append(bs, k)
	}

	return metrics.Parse(strings.Join(bs, "-"))
}
