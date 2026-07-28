package dao

import (
	"genshin-quiz/generated/db/genshinquiz/public/model"
	"genshin-quiz/generated/oapi"
)

type SimplePoll struct {
	Poll model.Polls
	User model.Users
}

type DetailedPoll struct {
	Poll               model.Polls
	User               model.Users
	Translation        []model.PollTranslations
	Options            []model.PollOptions
	OptionTranslations map[int64]oapi.LocalizedText
	SubmissionCount    int64
}

type PollListParams struct {
	Page       int            // 页码，从1开始
	NumPerPage int            // 每页数量
	IsPublic   *bool          // 是否公开
	IsVoted    *bool          // 是否已经投过
	Author     *int64         // 创建者
	Category   *oapi.Category // 分类过滤，空字符串表示不过滤
	Type       string         // 类型筛选: all, available, expired
	Query      *string        // 关键字搜索
	Language   *[]string      // 支持语言，默认 'zh-CN'
	SortBy     string         // 排序方式
	SortDesc   bool           // 是否降序排列
}

type PollListResult struct {
	Polls []SimplePoll
	Total int
}
