package main

import "fmt"

type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("字段 %s 校验失败: %s", e.Field, e.Msg)
}

type Report struct {
	Title string
	Score int
}

func validate(r Report) error {
	if r.Title == "" {
		return &ValidationError{Field: "title", Msg: "不能为空"}
	}
	if r.Score < 0 || r.Score > 100 {
		return &ValidationError{Field: "score", Msg: "必须在 0~100"}
	}
	return nil // 直接返回字面量 nil，而不是 nil 指针
}

func submit(r Report) {
	if err := validate(r); err != nil {
		fmt.Println("提交被拒绝:", err.Error())
		return
	}
	fmt.Printf("提交成功: %+v\n", r)
}

func main() {
	submit(Report{Title: "期中填报", Score: 120})
	submit(Report{Title: "期末填报", Score: 95})
}
