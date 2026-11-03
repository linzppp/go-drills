package main

import "fmt"

type Summary struct {
	Name   string
	Scores map[string]int
}

func NewSummary(name string) *Summary {
	return &Summary{Name: name, Scores: map[string]int{}}
}

func (s *Summary) Add(subject string, score int) {
	s.Scores[subject] = score
}

func main() {
	s := NewSummary("高一(3)班")
	fmt.Println("读 nil map:", s.Scores["数学"]) // 读取是安全的
	s.Add("数学", 90)
	fmt.Println(s)
}
