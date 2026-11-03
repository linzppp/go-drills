package main

import "fmt"

type Conn struct{ name string }

func (c *Conn) Close() { fmt.Println("关闭连接", c.name) }

func open(name string) *Conn {
	fmt.Println("打开连接", name)
	return &Conn{name: name}
}

// 依次处理多个数据源，期望：处理完一个就释放一个
func processAll(names []string) {
	for _, n := range names {
		c := open(n)
		defer c.Close()
		fmt.Println("  处理", n)
	}
	fmt.Println("全部处理完毕")
}

func main() {
	processAll([]string{"A库", "B库", "C库"})
}
