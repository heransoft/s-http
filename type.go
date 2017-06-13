package s_http

const (
	mainChanLength = 1024
)

type serverEvent struct {
	i interface{}
	e error
	c chan interface{}
}

type clientEvent struct {
	i interface{}
	e error
	f func(interface{}, error)
}
