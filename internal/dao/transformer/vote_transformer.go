package transformer

import (
	"time"

	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/dao"
)

// 将SimpleVote（不包含选项详情）模型转换为 OAPI.
func ConvertSimpleVoteToDTO(
	vote dao.SimpleVote,
) oapi.Vote {
	// 判断是否已过期
	expired := vote.Vote.ExpiresAt != nil && vote.Vote.ExpiresAt.Before(time.Now())

	// 构建多语言标题和描述（简单模式只有一种语言）
	title := make(map[string]string)
	description := make(map[string]string)
	title[vote.Translation.Language] = vote.Translation.Title
	if vote.Translation.Description != nil {
		description[vote.Translation.Language] = *vote.Translation.Description
	}

	participants := int(vote.Vote.ParticipantsCount)
	totalVotes := int(vote.Vote.TotalVotesCount)
	likes := int(vote.Vote.LikesCount)
	votesPerUser := int(vote.Vote.VotesPerUser)

	var votesPerOption *int
	if vote.Vote.VotesPerOption != nil {
		val := int(*vote.Vote.VotesPerOption)
		votesPerOption = &val
	}

	var expireAt *time.Time
	if vote.Vote.ExpiresAt != nil {
		expireAt = vote.Vote.ExpiresAt
	}

	return oapi.Vote{
		Id:             vote.Vote.VoteUUID,
		Public:         vote.Vote.Public,
		Title:          title,
		Description:    &description,
		Category:       oapi.QuestionCategory(vote.Vote.Category),
		StartAt:        vote.Vote.StartAt,
		ExpireAt:       expireAt,
		Expired:        expired,
		VotedOptions:   nil, // 简单模式不返回
		VotesPerUser:   votesPerUser,
		VotesPerOption: *votesPerOption,
		Options:        nil, // 简单模式不返回选项
		Participants:   &participants,
		TotalVotes:     &totalVotes,
		CreatedBy:      vote.User.UserUUID,
		CreatedAt:      vote.Vote.CreatedAt,
		Likes:          &likes,
	}
}

func ConvertDetailedVoteToDTO(
	vote dao.DetailedVote,
	votedOptions []oapi.VoteSubmissionOption,
	likeStatus int16,
) oapi.Vote {
	// 判断是否已过期
	expired := vote.Vote.ExpiresAt != nil && vote.Vote.ExpiresAt.Before(time.Now())

	// 构建多语言标题和描述
	title := make(map[string]string)
	description := make(map[string]string)

	// 检查是否有翻译数据
	if vote.Translation.Language != "" {
		title[vote.Translation.Language] = vote.Translation.Title
		if vote.Translation.Description != nil {
			description[vote.Translation.Language] = *vote.Translation.Description
		}
	}

	// 构建选项翻译的映射：optionID -> language -> text
	optionTranslationsMap := make(map[int64]map[string]string)
	for _, trans := range vote.OptionTranslations {
		if _, exists := optionTranslationsMap[trans.OptionID]; !exists {
			optionTranslationsMap[trans.OptionID] = make(map[string]string)
		}
		optionTranslationsMap[trans.OptionID][trans.Language] = trans.OptionText
	}

	// 转换选项
	options := make([]oapi.VoteOption, 0, len(vote.Options))
	for _, opt := range vote.Options {
		votes := int(opt.VoteCount)
		optionUUID := opt.OptionUUID

		// 获取该选项的多语言文本
		var optionText *map[string]string
		if texts, ok := optionTranslationsMap[opt.ID]; ok && len(texts) > 0 {
			optionText = &texts
		}

		options = append(options, oapi.VoteOption{
			Id:    &optionUUID,
			Text:  optionText,
			Type:  oapi.VoteOptionType("text"), // 默认文本类型，根据实际情况调整
			Votes: &votes,
		})
	}

	participants := int(vote.Vote.ParticipantsCount)
	totalVotes := int(vote.Vote.TotalVotesCount)
	likes := int(vote.Vote.LikesCount)
	votesPerUser := int(vote.Vote.VotesPerUser)

	var votesPerOption *int
	if vote.Vote.VotesPerOption != nil {
		val := int(*vote.Vote.VotesPerOption)
		votesPerOption = &val
	} else {
		zero := 0
		votesPerOption = &zero
	}

	var expireAt *time.Time
	if vote.Vote.ExpiresAt != nil {
		expireAt = vote.Vote.ExpiresAt
	}

	return oapi.Vote{
		Id:             vote.Vote.VoteUUID,
		Public:         vote.Vote.Public,
		Title:          title,
		Description:    &description,
		Category:       oapi.QuestionCategory(vote.Vote.Category),
		StartAt:        vote.Vote.StartAt,
		ExpireAt:       expireAt,
		Expired:        expired,
		VotedOptions:   votedOptions,
		VotesPerUser:   votesPerUser,
		VotesPerOption: *votesPerOption,
		Options:        options,
		Participants:   &participants,
		TotalVotes:     &totalVotes,
		CreatedBy:      vote.User.UserUUID,
		CreatedAt:      vote.Vote.CreatedAt,
		Likes:          &likes,
	}
}
