package monitor

type IClient interface {
	ID() string
	Write(point IPoint) error
	Close()
}
