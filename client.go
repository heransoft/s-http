package s_http

import (
	"net/http"
)

type Client struct {
	mainChan chan *clientEvent

	marshal   func(interface{}) *http.Request
	unmarshal func(*http.Response, error) interface{}
}

func NewClient(
	marshal func(interface{}) *http.Request,
	unmarshal func(*http.Response, error) interface{},
) *Client {
	p := new(Client)
	p.mainChan = make(chan *clientEvent, mainChanLength)

	p.marshal = marshal
	p.unmarshal = unmarshal
	return p
}

func (p *Client) GetMainChan() <-chan *clientEvent {
	return p.mainChan
}

func (p *Client) Deal(e *clientEvent) {
	e.f(e.i, e.e)
}

func (p *Client) Send(data interface{}, onMessage func(interface{}, error)) {
	go func() {
		req := p.marshal(data)
		c := &http.Client{}
		res, e := c.Do(req)
		i := p.unmarshal(res, e)
		p.mainChan <- &clientEvent{
			i: i,
			e: e,
			f: onMessage,
		}
	}()
}
