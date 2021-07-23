package checker

type CheckType int

const (
	CheckTypeEqual CheckType = iota
	CheckTypePrefix
	CheckTypeSuffix
	CheckTypeNotEqual
	CheckTypeNull
	CheckTypeExist
	CheckTypeNotExist
	CheckTypeRegular
	CheckTypeRegularG
	CheckTypeAll
)
