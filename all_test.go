package s_http_test

import (
	"fmt"
	"github.com/heransoft/s-http"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
	"time"
	"github.com/heransoft/sync-http"
)

func TestClient_Send(t *testing.T) {
	main1ThreadExitChan := make(chan int64, 1)
	main1ThreadExitedChan := make(chan int64, 1)

	go func() {
		s := s_http.NewServer(
			func(r *http.Request) (interface{}, error) {
				v := r.FormValue("p")
				return v, nil
			},
			func(v interface{}, e error) interface{} {
				return v
			},
			func(v interface{}) []byte {
				s := v.(string)
				return []byte(s)
			},
		)
		go func() {
			r := int64(0)
			defer func() {
				main1ThreadExitedChan <- r
			}()
			for {
				select {
				case result := <-main1ThreadExitChan:
					r = result
					return
				case mainChanElement := <-s.GetMainChan():
					s.Deal(mainChanElement)
				}
			}
		}()
		mux := http.NewServeMux()
		mux.HandleFunc("/k", s.HandleFunc)
		http.ListenAndServe(":8089", mux)
	}()
	time.Sleep(time.Second)

	main2ThreadExitChan := make(chan int64, 1)
	main2ThreadExitedChan := make(chan int64, 1)
	caseCount := int32(2000)
	caseThreadExitedChan := make(chan int64, caseCount)
	go func() {
		c := sync_http.NewClient(
			func(d interface{}) *http.Request {
				urlStr := fmt.Sprintf("http://localhost:8089/k?p=%s", d)
				r, e := http.NewRequest("GET", urlStr, nil)
				if e != nil {
					panic(e)
				}
				return r
			},
			func(r *http.Response, e error) interface{} {
				defer r.Body.Close()
				body, _ := ioutil.ReadAll(r.Body)
				return fmt.Sprintf("%s", body)
			},
		)
		go func() {
			r := int64(0)
			defer func() {
				main2ThreadExitedChan <- r
			}()
			for {
				select {
				case result := <-main2ThreadExitChan:
					r = result
					return
				case mainChanElement := <-c.GetMainChan():
					c.Deal(mainChanElement)
				}
			}
		}()
		for i := int32(0); i < caseCount; i++ {
			send := fmt.Sprint(i)
			c.Send(send, func(d interface{}, e error) {
				result := fmt.Sprint(reflect.TypeOf(d), d, e)
				if result != "string"+send+"<nil>" {
					t.Error("send error")
				}
				caseThreadExitedChan <- 0
			})
		}

	}()

	caseThreadExitedCount := int32(0)
	for {
		<-caseThreadExitedChan
		caseThreadExitedCount++
		if caseCount == caseThreadExitedCount {
			main1ThreadExitChan <- 0
			main2ThreadExitChan <- 0
			<-main1ThreadExitedChan
			<-main2ThreadExitedChan
			break
		}
	}
}
