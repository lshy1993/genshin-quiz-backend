package dao

import (
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/oapi"
	"time"
)

type SimpleQuestion struct {
	Question    model.Questions
	User        model.Users
	Translation model.QuestionTranslations
	// Solved      bool
}

type DetailedQuestion struct {
	Question           model.Questions
	User               model.Users
	Translation        model.QuestionTranslations
	Options            []model.QuestionOptions
	OptionTranslations []model.OptionTranslations
	SubmissionCount    int64
}

type QuestionListParams struct {
	Page       int                        // 页码，从1开始
	NumPerPage int                        // 每页数量
	Category   *oapi.QuestionCategory     // 分类过滤，空字符串表示不过滤
	Difficulty *[]oapi.QuestionDifficulty // 难度过滤，空字符串表示不过滤
	Query      *string                    // 关键字搜索，空字符串表示不搜索
	SortBy     *string                    // 排序方式
	SortDesc   bool                       // 是否降序排列，默认false（升序）
	Language   *[]string                  // 语言，默认 'zh-CN'
}

type QuestionListResult struct {
	Questions []SimpleQuestion
	Total     int
}

type SubmissionWithUserName struct {
	ID        int64
	IsCorrect bool
	CreatedAt time.Time
	TimeTaken *int32
	UserName  string
}

// const (
// 	SortByDate       SortBy = "date"
// 	SortByDifficulty SortBy = "difficulty"
// 	SortByLikes      SortBy = "likes"
// 	SortByText       SortBy = "text"
// )
