package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"time"
)

func ping(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) // 读完才能复用连接
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}

func openFDs() int {
	ents, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		return -1
	}
	return len(ents)
}

func main() {
	time.AfterFunc(5*time.Minute, func() { os.Exit(0) })
	go func() { log.Println(http.ListenAndServe("localhost:6060", nil)) }()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})
	go func() { log.Println(http.ListenAndServe("localhost:8081", mux)) }()
	time.Sleep(200 * time.Millisecond)

	go func() {
		for {
			if err := ping("http://localhost:8081/health"); err != nil {
				log.Println("ping failed:", err)
			}
			time.Sleep(20 * time.Millisecond)
		}
	}()

	for {
		time.Sleep(2 * time.Second)
		fmt.Printf("[%s] goroutines=%d fds=%d\n", time.Now().Format("15:04:05"), runtime.NumGoroutine(), openFDs())
	}
}
