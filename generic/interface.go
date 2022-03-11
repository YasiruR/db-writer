package generic

type Database interface {
	Init(addr, pw string)
	Write()
}
