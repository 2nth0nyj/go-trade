package core

type Account interface {
	Balance() string
	Broker() string
}
