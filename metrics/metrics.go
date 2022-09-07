package metrics

import (
	"strings"
)

type LabelReader interface {
	GetLabel(name string) string
}
type metricsReader interface {
	reader(labels LabelReader) string
}
type Metrics interface {
	Metrics(ctx LabelReader) string
}

func Parse(metrics []string) Metrics {

	ms := make(metricsList, 0, len(metrics))

	for _, k := range metrics {
		l := len(k)
		if l == 0 {
			continue
		}
		if len(k) > 2 {
			if k[0] == '{' && k[l-1] == '}' {
				r := k[1 : l-1]
				if len(r) == 0 {
					continue
				}
				ms = append(ms, metricsLabelReader(r))
				continue
			}
		}
		ms = append(ms, metricsConst(k))
	}

	return ms
}

type metricsLabelReader string

func (m metricsLabelReader) reader(labels LabelReader) string {
	return labels.GetLabel(string(m))
}

type metricsConst string

func (m metricsConst) reader(labels LabelReader) string {
	return string(m)
}

type metricsList []metricsReader

func (ms metricsList) Metrics(ctx LabelReader) string {
	vs := make([]string, len(ms))
	for i, r := range ms {
		vs[i] = r.reader(ctx)
	}
	return strings.Join(vs, "-")
}
