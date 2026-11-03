package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"time"
)

var recent [][]byte

// 读取一份 2MB 的"报表原文"，只保留前 64 字节作为摘要
func summarize(id int) []byte {
	raw := make([]byte, 2<<20)
	for i := range raw {
		raw[i] = byte(id + i)
	}
	return raw[:64]
}

func main() {
	time.AfterFunc(2*time.Minute, func() { os.Exit(0) })
	go func() { log.Println(http.ListenAndServe("localhost:6060", nil)) }()

	go func() {
		for i := 0; ; i++ {
			recent = append(recent, summarize(i))
			time.Sleep(200 * time.Millisecond)
		}
	}()

	for {
		time.Sleep(2 * time.Second)
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Printf("[%s] summaries=%d heapInuse=%dMB\n", time.Now().Format("15:04:05"), len(recent), m.HeapInuse>>20)
	}
}
