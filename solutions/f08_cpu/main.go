package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"regexp"
	"runtime"
	"sync/atomic"
	"time"
)

var phoneRe = regexp.MustCompile(`^1[3-9]\d{9}$`) // 只编译一次

func validatePhone(s string) bool {
	return phoneRe.MatchString(s)
}

func validateRows(rows []string) int {
	bad := 0
	for _, r := range rows {
		if !validatePhone(r) {
			bad++
		}
	}
	return bad
}

func main() {
	time.AfterFunc(5*time.Minute, func() { os.Exit(0) })
	go func() { log.Println(http.ListenAndServe("localhost:6060", nil)) }()

	rows := make([]string, 1000)
	for i := range rows {
		rows[i] = fmt.Sprintf("138%08d", i)
	}

	var done atomic.Int64
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for {
				validateRows(rows)
				done.Add(1)
			}
		}()
	}

	for {
		time.Sleep(2 * time.Second)
		fmt.Printf("[%s] batches/s=%d (每批 1000 行)\n", time.Now().Format("15:04:05"), done.Swap(0)/2)
	}
}
