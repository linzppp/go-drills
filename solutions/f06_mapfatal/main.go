package main

import (
	"fmt"
	"sync"
	"time"
)

type Stats struct {
	mu     sync.Mutex
	counts map[string]int
}

func (s *Stats) Inc(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counts[key]++
}

func safeInc(s *Stats, key string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered:", r)
		}
	}()
	s.Inc(key)
}

func main() {
	s := &Stats{counts: make(map[string]int)}
	keys := []string{"school-a", "school-b", "school-c", "school-d"}

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 200000; j++ {
				safeInc(s, keys[(id+j)%len(keys)])
			}
		}(i)
	}
	wg.Wait()
	time.Sleep(100 * time.Millisecond)
	fmt.Println("final:", s.counts)
}
