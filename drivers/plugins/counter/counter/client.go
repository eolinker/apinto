package counter

type IClient interface {
	Get(key string) (int64, error)
}
