package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"sync/atomic"
	"time"
)

type Row struct {
	ID    int
	Name  string
	Score float64
}

func exportCSV(rows []Row) []byte {
	out := "id,name,score\n"
	for _, r := range rows {
		line := fmt.Sprintf("%d,%s,%.2f", r.ID, r.Name, r.Score)
		out += line + "\n"
	}
	return []byte(out)
}

func main() {
	time.AfterFunc(5*time.Minute, func() { os.Exit(0) })
	go func() { log.Println(http.ListenAndServe("localhost:6060", nil)) }()

	rows := make([]Row, 3000)
	for i := range rows {
		rows[i] = Row{ID: i, Name: fmt.Sprintf("student-%d", i), Score: float64(i%100) + 0.5}
	}

	var done atomic.Int64
	for i := 0; i < 4; i++ {
		go func() {
			for {
				exportCSV(rows)
				done.Add(1)
			}
		}()
	}

	var last runtime.MemStats
	for {
		time.Sleep(2 * time.Second)
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Printf("[%s] exports/s=%d gcCycles(+%d) allocated=%dMB/s\n",
			time.Now().Format("15:04:05"), done.Swap(0)/2, m.NumGC-last.NumGC, (m.TotalAlloc-last.TotalAlloc)>>21)
		last = m
	}
}
