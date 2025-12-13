package dao

import (
	"genshin-quiz/generated/db/genshinquiz/public/model"
)

type SimpleVote struct {
	Vote        model.Votes
	Translation model.VoteTranslations
	User        model.Users
}

type DetailedVote struct {
	Vote               model.Votes
	User               model.Users
	Translation        model.VoteTranslations
	Options            []model.VoteOptions
	OptionTranslations []model.VoteOptionTranslations
	SubmissionCount    int64
}

type VoteListParams struct {
	Page     int       // 页码，从1开始
	Limit    int       // 每页数量
	Type     string    // 类型筛选: all, available, expired
	Query    *string   // 关键字搜索
	Language *[]string // 语言
	SortBy   string    // 排序方式
	SortDesc bool      // 是否降序排列
}

type VoteListResult struct {
	Votes []SimpleVote
	Total int
}
