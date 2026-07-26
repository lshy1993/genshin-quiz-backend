package transformer

import (
	"genshin-quiz/generated/oapi"
	"genshin-quiz/internal/dao"
)

// 将SimplePoll（不包含选项详情）模型转换为 OAPI.
func ConvertSimplePollToDTO(
	poll dao.SimplePoll,
	voted bool,
	likeStatus int16,
) oapi.Poll {
	// 构建多语言标题和描述（简单模式只有一种语言）
	title := make(map[string]string)
	description := make(oapi.LocalizedText)
	title[poll.Translation.Language] = poll.Translation.Title
	if poll.Translation.Description != nil {
		description[poll.Translation.Language] = *poll.Translation.Description
	}

	participants := int(poll.Poll.ParticipantsCount)
	totalVotes := int(poll.Poll.TotalVotesCount)
	likes := int(poll.Poll.LikesCount)
	votesPerUser := int(poll.Poll.VotesPerUser)

	votesPerOption := int(poll.Poll.VotesPerOption)

	expireAt := poll.Poll.ExpiresAt

	likeStatusValue := oapi.LikeStatus(likeStatus)

	return oapi.Poll{
		Id:                poll.Poll.PollUUID,
		Public:            poll.Poll.Public,
		Title:             title,
		Description:       &description,
		Category:          oapi.Category(poll.Poll.Category),
		StartAt:           poll.Poll.StartAt,
		ExpireAt:          expireAt,
		MyVotes:           nil, // 简单模式不返回
		VotesPerUser:      votesPerUser,
		VotesPerOption:    votesPerOption,
		Options:           nil, // 简单模式不返回选项
		ParticipantsCount: participants,
		TotalVotesCount:   totalVotes,
		CreatedBy:         poll.User.UserUUID,
		CreatedAt:         poll.Poll.CreatedAt,
		LikesCount:        likes,
		LikeStatus:        likeStatusValue,
	}
}

func ConvertDetailedVoteToDTO(
	vote dao.DetailedPoll,
	votedOptions []oapi.PollVote,
	likeStatus int16,
) oapi.Poll {
	// 构建多语言标题和描述
	title := make(oapi.LocalizedText)
	description := make(oapi.LocalizedText)

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
	options := make([]oapi.PollOption, 0, len(vote.Options))
	for _, opt := range vote.Options {
		votes := int(opt.VoteCount)
		optionUUID := opt.OptionUUID

		// 获取该选项的多语言文本
		var optionText oapi.LocalizedText
		if texts, ok := optionTranslationsMap[opt.ID]; ok && len(texts) > 0 {
			optionText = texts
		}

		options = append(options, oapi.PollOption{
			Id:         optionUUID,
			Text:       optionText,
			OptionType: oapi.OptionType("text"), // 默认文本类型，根据实际情况调整
			VotesCount: votes,
		})
	}

	participants := int(vote.Poll.ParticipantsCount)
	totalVotes := int(vote.Poll.TotalVotesCount)
	likes := int(vote.Poll.LikesCount)
	votesPerUser := int(vote.Poll.VotesPerUser)

	votesPerOption := int(vote.Poll.VotesPerOption)
	expireAt := vote.Poll.ExpiresAt

	likeStatusValue := oapi.LikeStatus(likeStatus)

	return oapi.Poll{
		Id:                vote.Poll.PollUUID,
		Public:            vote.Poll.Public,
		Title:             title,
		Description:       &description,
		Category:          oapi.Category(vote.Poll.Category),
		StartAt:           vote.Poll.StartAt,
		ExpireAt:          expireAt,
		MyVotes:           votedOptions,
		VotesPerUser:      votesPerUser,
		VotesPerOption:    votesPerOption,
		Options:           options,
		ParticipantsCount: participants,
		TotalVotesCount:   totalVotes,
		CreatedBy:         vote.User.UserUUID,
		CreatedAt:         vote.Poll.CreatedAt,
		LikesCount:        likes,
		LikeStatus:        likeStatusValue,
	}
}
