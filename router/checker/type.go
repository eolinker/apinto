package checker

type CheckType int

const (
	CheckTypeEqual CheckType = iota
	CheckTypePrefix
	CheckTypeSuffix
	CheckTypeNotEqual
	CheckTypeExist
	CheckTypeNotExist
	CheckTypeNull
	CheckTypeRegular
	CheckTypeRegularG
	CheckTypeAll
)
