package model

import "time"

type Server struct {
	Id    int64
	Name  string
	Age   int
	Email string
}

type ComplexServer struct {
	Addr   string
	Time   time.Time
	Server Server
}
