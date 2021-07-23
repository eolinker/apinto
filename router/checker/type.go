package checker

type CheckType int

const (
	CheckTypeEqual CheckType = iota
	CheckTypePrefix
	CheckTypeSuffix
	CheckTypeNotEqual
	CheckTypeNone
	CheckTypeExist
	CheckTypeNotExist
	CheckTypeRegular
	CheckTypeRegularG
	CheckTypeAll
)
