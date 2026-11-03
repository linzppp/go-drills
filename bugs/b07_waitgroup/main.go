package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := []int{}

	for i := 1; i <= 100; i++ {
		go func() {
			wg.Add(1)
			defer wg.Done()
			mu.Lock()
			results = append(results, i*i)
			mu.Unlock()
		}()
	}
	wg.Wait()

	mu.Lock()
	fmt.Printf("期望 100 条结果，实际 %d 条\n", len(results))
	mu.Unlock()
}
