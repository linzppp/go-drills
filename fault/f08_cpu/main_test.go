package main

import (
	"fmt"
	"testing"
)

func BenchmarkValidateRows(b *testing.B) {
	rows := make([]string, 1000)
	for i := range rows {
		rows[i] = fmt.Sprintf("138%08d", i)
	}
	b.ReportAllocs()
	for b.Loop() {
		validateRows(rows)
	}
}
