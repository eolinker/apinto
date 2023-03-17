package checker

import "sort"

type Handler interface {
	Check(v string, has bool) bool
}

type multipleChecker struct {
	equals map[string]bool
	other  listChecker
}
type listChecker []Checker

func (ls listChecker) Len() int {
	return len(ls)
}

func (ls listChecker) Less(i, j int) bool {
	li, lj := ls[i], ls[j]
	if li.CheckType() != lj.CheckType() {
		return li.CheckType() < lj.CheckType()
	}
	return li.Value() < lj.Value()
}

func (ls listChecker) Swap(i, j int) {
	ls[i], ls[j] = ls[j], ls[i]
}

func (ls listChecker) Check(v string, has bool) bool {
	for _, c := range ls {
		if c.Check(v, has) {
			return true
		}
	}
	return false
}

func (m *multipleChecker) Check(v string, has bool) bool {
	if has && m.equals != nil {
		//全选逻辑处理
		for k, _ := range m.equals {
			if k == "ALL" {
				return true
			}
		}

		if ok := m.equals[v]; ok {
			return true
		}
	}
	return m.other.Check(v, has)
}

func NewMultipleChecker(checkers []Checker) Handler {
	other := make(listChecker, 0, len(checkers))
	equals := make(map[string]bool, len(checkers))
	for _, c := range checkers {
		if c.CheckType() == CheckTypeEqual {
			equals[c.Value()] = true
		} else {
			other = append(other, c)
		}
	}
	sort.Sort(other)
	return &multipleChecker{other: other, equals: equals}
}
