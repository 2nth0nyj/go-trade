package core

type Strategy interface {
	Init()
	OnData()
	OnEvent()
	Stop()
}
