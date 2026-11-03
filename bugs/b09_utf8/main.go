package main

import "fmt"

// 列表页只显示前 6 个字符，超出的用省略号
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func main() {
	title := "北京市海淀区中关村第三小学期中填报"
	fmt.Println("截断结果:", truncate(title, 6))
	fmt.Println("len:", len(title), " 字符数应为 17")

	// 校验"名称不超过 10 个字"
	name := "上海市浦东新区实验学校附属小学"
	if len(name) > 10 {
		fmt.Println("校验失败：名称过长")
	} else {
		fmt.Println("校验通过")
	}
}
