package dao

import (
	"genshin-quiz/generated/db/genshinquiz/public/model"
)

type SimplePoll struct {
	Poll        model.Polls
	Translation model.PollTranslations
	User        model.Users
}

type DetailedPoll struct {
	Poll               model.Polls
	User               model.Users
	Translation        model.PollTranslations
	Options            []model.PollOptions
	OptionTranslations []model.PollOptionTranslations
	SubmissionCount    int64
}

type PollListParams struct {
	Page     int       // 页码，从1开始
	Limit    int       // 每页数量
	Type     string    // 类型筛选: all, available, expired
	Query    *string   // 关键字搜索
	Language *[]string // 语言
	SortBy   string    // 排序方式
	SortDesc bool      // 是否降序排列
}

type PollListResult struct {
	Polls []SimplePoll
	Total int
}
