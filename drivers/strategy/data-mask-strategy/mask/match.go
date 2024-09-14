package mask

const (
	MatchInner    = "inner"
	MatchKeyword  = "keyword"
	MatchRegex    = "regex"
	MatchJsonPath = "json_path"
)

const (
	MatchInnerValueName     = "name"
	MatchInnerValuePhone    = "phone"
	MatchInnerValueIDCard   = "id-card"
	MatchInnerValueBankCard = "bank-card"
	MatchInnerValueDate     = "date"
	MatchInnerValueAmount   = "amount"
)

type IInnerMask interface {
	Exec(body []byte) ([]byte, error)
}
