package main

import "fmt"

type Student struct {
	Name  string
	Score int
}

// 给所有不及格的学生加 10 分
func curve(students []Student) {
	for i, s := range students {
		if s.Score < 60 {
			s.Score += 10
			students[i] = s
		}
	}
}

func main() {
	list := []Student{{"张三", 55}, {"李四", 80}, {"王五", 58}}
	curve(list)
	fmt.Println(list)

	// 再看一个 map 版本
	byName := map[string]Student{"张三": {"张三", 55}}
	// byName["张三"].Score = 65 // 取消注释，看编译器怎么说
	fmt.Println(byName)
}
