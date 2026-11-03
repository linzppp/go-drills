package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Report struct {
	Title    string    `json:"title"`
	CreateAt time.Time `json:"createAt"`
}

func main() {
	t := time.Date(2026, 10, 15, 9, 30, 5, 0, time.UTC)

	fmt.Println("用 Java 习惯格式化:", t.Format("yyyy-MM-dd HH:mm:ss"))

	out, _ := json.Marshal(Report{Title: "期中填报", CreateAt: t})
	fmt.Println("默认 JSON 输出:", string(out))

	// 前端传来 "2026-10-15 09:30:05"
	var r Report
	err := json.Unmarshal([]byte(`{"title":"x","createAt":"2026-10-15 09:30:05"}`), &r)
	fmt.Println("解析前端时间:", err)

	// 解析不带时区的字符串，时区是什么？
	p, _ := time.Parse("2006-01-02 15:04:05", "2026-10-15 09:30:05")
	fmt.Println("time.Parse 得到的时区:", p.Location())
}
