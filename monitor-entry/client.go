package monitor_entry

type IClient interface {
	ID() string
	Write(point IPoint) error
	Close()
}
