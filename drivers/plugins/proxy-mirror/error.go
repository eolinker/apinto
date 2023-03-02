package proxy_mirror

import "github.com/pkg/errors"

var (
	errUnsupportedContextType = errors.New("send mirror proxy fail. Unsupported Context Type")
	errHostNull               = errors.New("host can't be null when pass_host is rewrite. ")
	errUnsupportedPassHost    = errors.New("unsupported pass_host. ")
	errTimeout                = errors.New("timeout can't be smaller than 0. ")

	errRandomRangeNum = errors.New("random_range should be bigger than 0. ")
	errRandomPivotNum = errors.New("random_pivot should be bigger than 0. ")
	errRandomPivot    = errors.New("random_pivot should be smaller than random_range. ")

	errAddr = errors.New("addr is illegal. ")
)
