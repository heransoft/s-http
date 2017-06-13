package s_http

import (
	"net/http"
)

type Server struct {
	mainChan chan *serverEvent

	unmarshal func(r *http.Request) (interface{}, error)
	onMessage func(interface{}, error) interface{}
	marshal   func(interface{}) []byte
}

func NewServer(
	unmarshal func(r *http.Request) (interface{}, error),
	onMessage func(interface{}, error) interface{},
	marshal func(interface{}) []byte,
) *Server {
	p := new(Server)
	p.mainChan = make(chan *serverEvent, mainChanLength)

	p.unmarshal = unmarshal
	p.onMessage = onMessage
	p.marshal = marshal
	return p
}

func (p *Server) GetMainChan() <-chan *serverEvent {
	return p.mainChan
}

func (p *Server) Deal(e *serverEvent) {
	o := p.onMessage(e.i, e.e)
	e.c <- o
}

func (p *Server) HandleFunc(w http.ResponseWriter, r *http.Request) {
	i, e := p.unmarshal(r)
	c := make(chan interface{})
	p.mainChan <- &serverEvent{
		i: i,
		e: e,
		c: c,
	}
	o := <-c
	b := p.marshal(o)
	w.Write(b)
}
