package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Notifier struct {
	Endpoint *string
}

func (n *Notifier) Send(msg string) {
	fmt.Println("通知发送到:", *n.Endpoint, msg)
}

var notifier = &Notifier{}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println("recovered in middleware:", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})
	mux.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		go notifier.Send("填报已提交")
		fmt.Fprint(w, "submitted")
	})
	go http.ListenAndServe("localhost:8083", recoverMiddleware(mux))
	time.Sleep(200 * time.Millisecond)

	for _, path := range []string{"/ping", "/submit", "/ping"} {
		resp, err := http.Get("http://localhost:8083" + path)
		if err != nil {
			fmt.Println(path, "请求失败:", err)
			continue
		}
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Println(path, "->", resp.StatusCode, string(b))
		time.Sleep(300 * time.Millisecond)
	}
	fmt.Println("全部请求完成，服务仍然存活")
}
