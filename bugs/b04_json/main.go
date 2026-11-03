package main

import (
	"encoding/json"
	"fmt"
)

type Report struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

func main() {
	// 1) 雪花 ID 返回给前端
	out, _ := json.Marshal(Report{ID: 1849302847562301441, Title: "期中填报"})
	fmt.Println("服务端输出:", string(out))
	// 前端 JS 的 JSON.parse 会把 id 读成 1849302847562301400，请求回来后就是另一条记录

	// 2) 服务端用 map[string]any 接收
	var m map[string]any
	json.Unmarshal([]byte(`{"id": 1849302847562301441}`), &m)
	id := m["id"]
	fmt.Printf("map 接收: %v 类型 %T\n", id, id)
	fmt.Printf("还原成整数: %d\n", int64(id.(float64)))
}
