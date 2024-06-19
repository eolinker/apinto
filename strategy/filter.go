package strategy

import (
	"github.com/eolinker/apinto/checker"
	"github.com/eolinker/eosc/eocontext"
)

type FilterConfig map[string][]string

func ParseFilter(config FilterConfig) (IFilter, error) {
	fs := make(Filters, 0, len(config))
	for name, patterns := range config {

		if len(patterns) == 0 {
			continue
		}
		cks := make([]checker.Checker, 0, len(patterns))

		for _, p := range patterns {
			if name == "ip" {
				c, err := newIPChecker(p)
				if err != nil {
					return nil, err
				}
				cks = append(cks, c)
				continue
			}
			c, err := checker.Parse(p)
			if err != nil {
				return nil, err
			}
			if c.CheckType() == checker.CheckTypeAll {
				cks = nil
				break
			}
			cks = append(cks, c)
		}

		if len(cks) != 0 {
			fs = append(fs, &FilterItem{
				Handler: checker.NewMultipleChecker(cks),
				name:    name,
			})
		}
	}

	return fs, nil
}

type FilterItem struct {
	checker.Handler
	name string
}
type Filters []*FilterItem

func (fs Filters) Check(ctx eocontext.EoContext) bool {
	vs := ctx.Labels()
	for _, f := range fs {
		v, has := vs[f.name]
		if !f.Check(v, has) {
			return false
		}
	}
	return true
}
