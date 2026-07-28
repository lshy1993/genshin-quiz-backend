package dao

import (
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/oapi"
	"time"

	"github.com/google/uuid"
)

type SimpleQuestion struct {
	Question model.Questions
	User     model.Users
}

type DetailedQuestion struct {
	Question           model.Questions
	User               model.Users
	Translation        []model.QuestionTranslations
	Options            []model.QuestionOptions
	OptionTranslations map[int64]oapi.LocalizedText
	SubmissionCount    int64
}

type QuestionListParams struct {
	Page        int                // 页码，从1开始
	NumPerPage  int                // 每页数量
	IsPublic    *bool              // 是否公开
	IsPublished *bool              // 是否已公开
	Author      *int64             // 创建者
	Category    *oapi.Category     // 分类过滤，空字符串表示不过滤
	Difficulty  *[]oapi.Difficulty // 难度过滤，空字符串表示不过滤
	Query       *string            // 关键字搜索，空字符串表示不搜索
	Language    *[]string          // 支持语言，默认 'zh-CN'
	SortBy      string             // 排序方式
	SortDesc    bool               // 是否降序排列，默认false（升序）
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
	UserID    uuid.UUID
}

// const (
// 	SortByDate       SortBy = "date"
// 	SortByDifficulty SortBy = "difficulty"
// 	SortByLikes      SortBy = "likes"
// 	SortByText       SortBy = "text"
// )
