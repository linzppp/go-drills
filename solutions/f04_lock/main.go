package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type entry struct {
	val      string
	loadedAt time.Time
}

type Cache struct {
	mu   sync.Mutex
	data map[int]entry
	ttl  time.Duration
}

func NewCache(ttl time.Duration) *Cache {
	return &Cache{data: make(map[int]entry), ttl: ttl}
}

func loadFromDB(key int) string {
	time.Sleep(20 * time.Millisecond) // 模拟一次数据库查询
	return fmt.Sprintf("value-%d", key)
}

func (c *Cache) Get(key int) string {
	c.mu.Lock()
	if e, ok := c.data[key]; ok && time.Since(e.loadedAt) < c.ttl {
		c.mu.Unlock()
		return e.val
	}
	c.mu.Unlock()
	v := loadFromDB(key) // 慢操作放在锁外；同 key 并发会重复回源，需要时再上 singleflight
	c.mu.Lock()
	c.data[key] = entry{val: v, loadedAt: time.Now()}
	c.mu.Unlock()
	return v
}

func main() {
	time.AfterFunc(5*time.Minute, func() { os.Exit(0) })
	runtime.SetMutexProfileFraction(1)
	runtime.SetBlockProfileRate(1)
	go func() { log.Println(http.ListenAndServe("localhost:6060", nil)) }()

	cache := NewCache(50 * time.Millisecond)
	var ops atomic.Int64
	for i := 0; i < 50; i++ {
		go func() {
			for {
				cache.Get(rand.Intn(10))
				ops.Add(1)
			}
		}()
	}

	for {
		time.Sleep(2 * time.Second)
		fmt.Printf("[%s] ops/s=%d goroutines=%d\n", time.Now().Format("15:04:05"), ops.Swap(0)/2, runtime.NumGoroutine())
	}
}
