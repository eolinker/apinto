package checker

type CheckType int

const (
	CheckTypeEqual CheckType = iota
	CheckTypePrefix
	CheckTypeSuffix
	CheckTypeSub
	CheckTypeNotEqual
	CheckTypeNone
	CheckTypeExist
	CheckTypeNotExist
	CheckTypeRegular
	CheckTypeRegularG
	CheckTypeAll
)
