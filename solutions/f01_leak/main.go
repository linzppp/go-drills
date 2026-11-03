package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"time"
)

func slowQuery(id int) string {
	time.Sleep(time.Duration(50+rand.Intn(300)) * time.Millisecond)
	return fmt.Sprintf("result-%d", id)
}

func search(id int) (string, error) {
	ch := make(chan string, 1) // 带缓冲：超时后协程仍能发送并退出
	go func() {
		ch <- slowQuery(id)
	}()
	select {
	case r := <-ch:
		return r, nil
	case <-time.After(200 * time.Millisecond):
		return "", fmt.Errorf("query %d timeout", id)
	}
}

func main() {
	time.AfterFunc(5*time.Minute, func() { os.Exit(0) })
	go func() { log.Println(http.ListenAndServe("localhost:6060", nil)) }()

	mux := http.NewServeMux()
	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		res, err := search(rand.Intn(1000))
		if err != nil {
			http.Error(w, err.Error(), http.StatusGatewayTimeout)
			return
		}
		fmt.Fprint(w, res)
	})
	go func() { log.Println(http.ListenAndServe("localhost:8080", mux)) }()
	time.Sleep(200 * time.Millisecond)

	// 内置流量发生器
	go func() {
		for {
			go func() {
				resp, err := http.Get("http://localhost:8080/search")
				if err == nil {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
				}
			}()
			time.Sleep(10 * time.Millisecond)
		}
	}()

	for {
		time.Sleep(2 * time.Second)
		fmt.Printf("[%s] goroutines=%d\n", time.Now().Format("15:04:05"), runtime.NumGoroutine())
	}
}
