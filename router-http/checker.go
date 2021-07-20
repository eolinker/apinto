package router

import "regexp"

type IChecker interface {
	Check(v string) bool
}

func parseChecker(r string) IChecker {
	// todo parse checker

	return nil
}

type Prefix string

type Regexp struct {
	rule string
	*regexp.Regexp
}
type Indistinct string
