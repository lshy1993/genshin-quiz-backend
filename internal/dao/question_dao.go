package dao

import (
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/oapi"
)

type SimpleQuestion struct {
	Question    model.Questions
	User        model.Users
	Translation model.QuestionTranslations
	Solved      bool
}

type DetailedQuestion struct {
	SimpleQuestion
	Submissions        model.QuestionSubmissions
	Options            []model.QuestionOptions
	OptionTranslations []model.OptionTranslations
}

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

type QuestionListResult struct {
	Questions []oapi.Question
	Total     int
}

// const (
// 	SortByDate       SortBy = "date"
// 	SortByDifficulty SortBy = "difficulty"
// 	SortByLikes      SortBy = "likes"
// 	SortByText       SortBy = "text"
// )
