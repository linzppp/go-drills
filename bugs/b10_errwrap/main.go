package main

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("record not found")

type NotFoundError struct{ Table string }

func (e *NotFoundError) Error() string { return "not found in " + e.Table }

func repoFind(id int) error {
	return ErrNotFound
}

func serviceGet(id int) error {
	if err := repoFind(id); err != nil {
		return fmt.Errorf("serviceGet(%d): %v", id, err)
	}
	return nil
}

func serviceGet2(id int) error {
	return fmt.Errorf("serviceGet2(%d): %w", id, &NotFoundError{Table: "report"})
}

func handler(id int) {
	err := serviceGet(id)
	if err == ErrNotFound {
		fmt.Println("404 记录不存在")
		return
	}
	fmt.Println("500 内部错误:", err)

	err = serviceGet2(id)
	if nf, ok := err.(*NotFoundError); ok {
		fmt.Println("404 记录不存在", nf.Table)
		return
	}
	fmt.Println("500 内部错误:", err)
}

func main() { handler(7) }
