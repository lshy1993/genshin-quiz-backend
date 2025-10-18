package dao

import (
	"genshin-quiz/generated/oapi"
)

// QuestionListParams 查询参数
type QuestionListParams struct {
	Page       int      // 页码，从1开始
	NumPerPage int      // 每页数量
	Category   string   // 分类过滤，空字符串表示不过滤
	Difficulty string   // 难度过滤，空字符串表示不过滤
	Query      string   // 关键字搜索，空字符串表示不搜索
	SortBy     string   // 排序方式
	SortDesc   bool     // 是否降序排列，默认false（升序）
	Language   []string // 语言，默认 'zh-CN'
}

// QuestionListResult 查询结果
type QuestionListResult struct {
	Questions []oapi.Question
	Total     int
}

// SortBy 排序方式
type SortBy string

const (
	SortByDate       SortBy = "date"
	SortByDifficulty SortBy = "difficulty"
	SortByLikes      SortBy = "likes"
	SortByText       SortBy = "text"
)
