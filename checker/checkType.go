package checker

// CheckType Checker类型
type CheckType int

const (
	//CheckTypeEqual 全等匹配Checker类型
	CheckTypeEqual CheckType = iota
	//CheckTypePrefix 前缀匹配Checker类型
	CheckTypePrefix
	//CheckTypeSuffix 后缀匹配Checker类型
	CheckTypeSuffix
	//CheckTypeSub 子串匹配Checker类型
	CheckTypeSub
	//CheckTypeNotEqual 非等匹配Checker类型
	CheckTypeNotEqual
	//CheckTypeNone 空值匹配Checker类型
	CheckTypeNone
	//CheckTypeExist 存在匹配Checker类型
	CheckTypeExist
	//CheckTypeNotExist 不存在匹配Checker类型
	CheckTypeNotExist
	//CheckTypeRegular 区分大小写的正则匹配Checker类型
	CheckTypeRegular
	//CheckTypeRegularG 不区分大小写的正则匹配Checker类型
	CheckTypeRegularG
	//CheckTypeAll 任意匹配Checker类型
	CheckTypeAll
	// CheckMultiple 复合匹配
	CheckMultiple
	CheckTypeIP
	CheckTypeTimeRange
)
