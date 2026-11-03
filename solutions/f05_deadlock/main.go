package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type Account struct {
	mu      sync.Mutex
	id      int
	balance int
}

func Transfer(from, to *Account, amount int) {
	first, second := from, to // 全局统一加锁顺序：按 id 从小到大
	if first.id > second.id {
		first, second = second, first
	}
	first.mu.Lock()
	defer first.mu.Unlock()
	time.Sleep(time.Millisecond) // 模拟校验耗时
	second.mu.Lock()
	defer second.mu.Unlock()
	from.balance -= amount
	to.balance += amount
}

func main() {
	time.AfterFunc(5*time.Minute, func() { os.Exit(0) })
	go func() { log.Println(http.ListenAndServe("localhost:6060", nil)) }()

	a := &Account{id: 1, balance: 100000}
	b := &Account{id: 2, balance: 100000}
	var done atomic.Int64

	go func() {
		for {
			Transfer(a, b, 1)
			done.Add(1)
		}
	}()
	go func() {
		for {
			Transfer(b, a, 1)
			done.Add(1)
		}
	}()

	for {
		time.Sleep(2 * time.Second)
		fmt.Printf("[%s] transfers done=%d\n", time.Now().Format("15:04:05"), done.Load())
	}
}
