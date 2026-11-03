package main

import (
	"encoding/json"
	"fmt"
)

// 模拟数据库中的一行
type ReportRow struct {
	ID        int
	Title     string
	Score     int
	Published bool
}

// 前端 PATCH 请求体
type UpdateReportReq struct {
	Title     *string `json:"title"`
	Score     *int    `json:"score"`
	Published *bool   `json:"published"`
}

// 等价于 ORM 里的 db.Model(&row).Updates(fields)
func applyUpdates(row *ReportRow, fields map[string]any) {
	if v, ok := fields["title"]; ok {
		row.Title = v.(string)
	}
	if v, ok := fields["score"]; ok {
		row.Score = v.(int)
	}
	if v, ok := fields["published"]; ok {
		row.Published = v.(bool)
	}
}

func handleUpdate(row *ReportRow, body string) error {
	var req UpdateReportReq
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return err
	}
	fields := map[string]any{}
	if req.Title != nil { // nil 才表示"没传"
		fields["title"] = *req.Title
	}
	if req.Score != nil {
		fields["score"] = *req.Score
	}
	if req.Published != nil {
		fields["published"] = *req.Published
	}
	applyUpdates(row, fields)
	return nil
}

func main() {
	row := &ReportRow{ID: 1, Title: "期中填报", Score: 80, Published: true}
	fmt.Printf("更新前: %+v\n", *row)

	// 用户把分数改成 0，并取消发布
	if err := handleUpdate(row, `{"score": 0, "published": false}`); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("更新后: %+v   (期望 Score=0, Published=false)\n", *row)
}
