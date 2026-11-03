package main

import (
	"errors"
	"fmt"
	"strings"
)

// ---- 假事务，模拟数据库行为 ----
type Tx struct {
	rows      []string
	committed bool
}

func (t *Tx) Insert(row string) error {
	if strings.TrimSpace(row) == "" {
		return errors.New("empty row")
	}
	for _, r := range t.rows {
		if r == row {
			return fmt.Errorf("duplicate key %q", row)
		}
	}
	t.rows = append(t.rows, row)
	return nil
}
func (t *Tx) Commit() error   { t.committed = true; return nil }
func (t *Tx) Rollback() error { t.rows = nil; return nil }

// ---- 业务代码：批量导入填报数据 ----
func importRows(tx *Tx, rows []string) error {
	var err error
	for _, r := range rows {
		if err := tx.Insert(r); err != nil {
			err = fmt.Errorf("insert %q: %w", r, err)
			continue
		}
	}
	return err
}

func main() {
	tx := &Tx{}
	input := []string{"张三,90", "李四,85", "", "张三,90", "王五,77"}

	if err := importRows(tx, input); err != nil {
		tx.Rollback()
		fmt.Println("导入失败，已回滚:", err)
		return
	}
	tx.Commit()
	fmt.Printf("导入成功：输入 %d 行，入库 %d 行，committed=%v\n", len(input), len(tx.rows), tx.committed)
}
